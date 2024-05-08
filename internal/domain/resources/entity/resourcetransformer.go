package entity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
			return r.transform(key)
		})
	if err != nil {
		return err
	}

	r.Resource = *res
	return nil
}

func (r *ResourceTransformer) transform(key string) (*Resource, error) {
	cache := r.TransformationCache

	b1 := bp.GetBuffer()
	b2 := bp.GetBuffer()
	defer bp.PutBuffer(b1)
	defer bp.PutBuffer(b2)

	tctx := &valueobject.ResourceTransformationCtx{
		Ctx:                   context.Background(),
		Data:                  make(map[string]any),
		OpenResourcePublisher: r.Resource.openPublishFileForWriting,
		DependencyManager:     r.Resource.sd.DependencyManager,
	}

	tctx.InMediaType = r.Resource.MediaType()
	tctx.OutMediaType = r.Resource.MediaType()

	startCtx := *tctx
	updates := &valueobject.TransformationUpdate{StartCtx: startCtx}

	var contentrc pio.ReadSeekCloser

	contentrc, err := valueobject.ContentReadSeekerCloser(&r.Resource)
	if err != nil {
		return nil, err
	}

	defer contentrc.Close()

	tctx.From = contentrc
	tctx.To = b1

	tctx.InPath = r.Resource.paths.TargetPath()
	tctx.SourcePath = strings.TrimPrefix(tctx.InPath, "/")

	counter := 0
	writeToFileCache := false

	var transformedContentr io.Reader

	for i, tr := range r.transformations {
		if i != 0 {
			tctx.InMediaType = tctx.OutMediaType
		}

		mayBeCachedOnDisk := transformationsToCacheOnDisk[tr.Key().Name] // is css
		if !writeToFileCache {
			writeToFileCache = mayBeCachedOnDisk
		}

		if i > 0 {
			hasWrites := tctx.To.(*bytes.Buffer).Len() > 0
			if hasWrites {
				counter++
				// Switch the buffers
				if counter%2 == 0 {
					tctx.From = b2
					b1.Reset()
					tctx.To = b1
				} else {
					tctx.From = b1
					b2.Reset()
					tctx.To = b2
				}
			}
		}

		newErr := func(err error) error {
			msg := fmt.Sprintf("%s: failed to transform %q (%s)", strings.ToUpper(tr.Key().Name), tctx.InPath, tctx.InMediaType.Type)

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

		if err != nil {
			return nil, newErr(err)
		}

		if tctx.OutPath != "" {
			tctx.InPath = tctx.OutPath
			tctx.OutPath = ""
		}
	}

	if transformedContentr == nil {
		updates.UpdateFromCtx(tctx)
	}

	var publishwriters []io.WriteCloser

	publicw, err := r.Resource.openPublishFileForWriting(updates.TargetPath)
	if err != nil {
		return nil, err
	}
	publishwriters = append(publishwriters, publicw)

	if transformedContentr == nil {
		if writeToFileCache {
			// Also write it to the cache
			fi, metaw, err := cache.WriteMeta(key, updates.ToTransformedResourceMetadata())
			if err != nil {
				return nil, err
			}
			updates.SourceFilename = &fi.Name
			updates.SourceFs = cache.Caches.AssetsCache().Fs
			publishwriters = append(publishwriters, metaw)
		}

		// Any transformations reading from From must also write to To.
		// This means that if the target buffer is empty, we can just reuse
		// the original reader.
		if b, ok := tctx.To.(*bytes.Buffer); ok && b.Len() > 0 {
			transformedContentr = tctx.To.(*bytes.Buffer)
		} else {
			transformedContentr = contentrc
		}
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

	s := contentmemw.String()
	updates.Content = &s

	newTarget, err := r.Resource.cloneWithUpdates(updates)
	if err != nil {
		return nil, err
	}

	return newTarget, nil
}
