package entity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"io"
	"strings"
)

type ResourceTransformer struct {
	Resource

	publisher *Publisher
	mediaSvc  resources.MediaTypesConfig

	TransformationCache *Cache

	*resourceTransformations
}

func (r *ResourceTransformer) Transform(t ...ResourceTransformation) (ResourceTransformable, error) {
	return r.TransformWithContext(context.Background(), t...)
}

func (r *ResourceTransformer) TransformWithContext(ctx context.Context, t ...ResourceTransformation) (ResourceTransformable, error) {
	r.resourceTransformations = &resourceTransformations{
		transformations: append(r.transformations, t...),
	}

	r.startTransform()
	return r, nil
}

func (r *ResourceTransformer) startTransform() {
	r.transformationsInit.Do(func() {
		if len(r.transformations) == 0 {
			// Nothing to do.
			return
		}

		r.transformationsErr = r.getOrTransform()
		if r.transformationsErr != nil {
			_ = fmt.Errorf("transformation failed: %s", r.transformationsErr)
		}
	})
}

func (r *ResourceTransformer) TransformationKey() string {
	var key string
	for _, tr := range r.transformations {
		key = key + "_" + tr.Key().Value()
	}
	return r.TransformationCache.CleanKey(r.Resource.Key()) + "_" + helpers.MD5String(key)
}

func (r *ResourceTransformer) getOrTransform() error {
	key := r.TransformationKey()
	res, err := r.TransformationCache.CacheResourceTransformation.GetOrCreate(
		key, func(string) (*Resource, error) {
			res, err := r.getFromFile(key)
			if err != nil {
				return nil, err
			}

			if res != nil {
				return res, nil
			}

			return r.transform(key)
		})
	if err != nil {
		return err
	}

	r.Resource = *res
	return nil
}

func (r *ResourceTransformer) getFromFile(key string) (*Resource, error) {
	_, f, metaBytes, found := r.TransformationCache.GetFile(key)
	if !found {
		return nil, nil
	}

	meta, err := r.Resource.meta().Unmarshal(metaBytes)
	if err != nil {
		return nil, err
	}

	m, found := r.mediaSvc.LookByType(meta.MediaTypeV)
	if !found {
		return nil, errors.New("media type not found")
	}

	r2 := r.Resource.clone()

	r2.mediaType = m
	r2.paths = valueobject.NewResourcePaths(meta.Target)
	r2.mergeData(meta.MetaData)
	r2.openReadSeekCloser = func() (pio.ReadSeekCloser, error) {
		return f.(pio.ReadSeekCloser), nil
	}

	return r2, nil
}

func (r *ResourceTransformer) transform(key string) (*Resource, error) {
	cache := r.TransformationCache

	var contentrc pio.ReadSeekCloser
	contentrc, err := valueobject.ContentReadSeekerCloser(&r.Resource)
	if err != nil {
		return nil, err
	}
	defer contentrc.Close()

	ctxBuilder := valueobject.NewResourceTransformationCtxBuilder(&r.Resource, r.publisher).
		WithMediaType(r.Resource.mediaType).
		WithSource(contentrc).
		WithTargetPath(r.Resource.paths.TargetPath())
	tctx := ctxBuilder.Build()
	defer tctx.Close()

	for _, tr := range r.transformations {

		tctx.UpdateBuffer()

		newErr := func(err error) error {
			msg := fmt.Sprintf("%s: failed to transform %q (%s)", strings.ToUpper(tr.Key().Name), tctx.Source.InPath, tctx.Source.InMediaType.Type)

			if herrors.IsFeatureNotAvailableError(err) {
				var errMsg string
				if tr.Key().Name == "postcss" {
					// This transformation is not available in this
					// Most likely because PostCSS is not installed.
					errMsg = ". Check your PostCSS installation; install with \"npm install postcss-cli\". See https://gohugo.io/hugo-pipes/postcss/"
				} else if tr.Key().Name == "tocss" {
					errMsg = ". Check your Hugo installation; you need the extended version to build SCSS/SASS with transpiler set to 'libsass'."
				} else if tr.Key().Name == "tocss-dart" {
					errMsg = ". You need dart-sass-embedded in your system $PATH."
				} else if tr.Key().Name == "babel" {
					errMsg = ". You need to install Babel, see https://gohugo.io/hugo-pipes/babel/"
				}

				return fmt.Errorf(msg+errMsg+": %w", err)
			}

			return fmt.Errorf(msg+": %w", err)
		}

		err = tr.Transform(tctx)
		if err != nil && !errors.Is(err, herrors.ErrFeatureNotAvailable) {
			return nil, newErr(err)
		}

		tctx.UpdateSource()
	}

	updates := r.Resource.clone()

	updates.mediaType = tctx.Source.InMediaType
	updates.data = tctx.Data
	updates.paths = valueobject.NewResourcePaths(tctx.Source.InPath)

	var publishwriters []io.WriteCloser
	publicw, err := r.publisher.OpenPublishFileForWriting(updates.paths.TargetPath())
	if err != nil {
		return nil, err
	}
	publishwriters = append(publishwriters, publicw)

	// Also write it to the cache
	metaBytes, err := updates.meta().Marshal()
	if err != nil {
		return nil, err
	}
	_, file, err := cache.WriteMeta(key, metaBytes)
	if err != nil {
		return nil, err
	}
	publishwriters = append(publishwriters, file)

	var transformedContentr io.Reader
	if b, ok := tctx.Target.To.(*bytes.Buffer); ok && b.Len() > 0 {
		transformedContentr = tctx.Target.To.(*bytes.Buffer)
	} else {
		transformedContentr = contentrc
	}

	// Also write it to memory
	var contentmemw *bytes.Buffer

	contentmemw = bp.GetBuffer()
	defer bp.PutBuffer(contentmemw)
	publishwriters = append(publishwriters, pio.ToWriteCloser(contentmemw))

	publishw := pio.NewMultiWriteCloser(publishwriters...)
	_, err = io.Copy(publishw, transformedContentr)
	if err != nil {
		return nil, err
	}
	publishw.Close()

	content := contentmemw.String()
	updates.openReadSeekCloser = func() (pio.ReadSeekCloser, error) {
		return pio.NewReadSeekerNoOpCloserFromString(content), nil
	}

	return updates, nil
}
