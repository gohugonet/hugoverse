package entity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/constants"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/images/exif"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"image"
	"io"
	"strings"
)

func newResourceAdapter(rc *ResourceCache, spec *valueobject.Spec, lazyPublish bool, target transformableResource) *resourceAdapter {
	var po *valueobject.PublishOnce
	if lazyPublish {
		po = &valueobject.PublishOnce{}
	}
	return &resourceAdapter{
		ResourceCache:           rc,
		resourceTransformations: &resourceTransformations{},
		metaProvider:            target,
		ResourceAdapterInner: &ResourceAdapterInner{
			ctx:         context.Background(),
			spec:        spec,
			PublishOnce: po,
			target:      target,
			Staler:      &valueobject.AtomicStaler{},
		},
	}
}

type resourceAdapter struct {
	ResourceCache *ResourceCache

	commonResource
	*resourceTransformations
	*ResourceAdapterInner
	metaProvider resources.MetaProvider
}

type commonResource struct{}

type ResourceAdapterInner struct {
	// The context that started this transformation.
	ctx context.Context

	target transformableResource

	spec *valueobject.Spec

	stale.Staler

	// Handles publishing (to /public) if needed.
	*valueobject.PublishOnce
}

func (r *ResourceAdapterInner) IsStale() bool {
	return r.Staler.IsStale() || r.target.IsStale()
}

var _ identity.ForEeachIdentityByNameProvider = (*resourceAdapter)(nil)

func (r *resourceAdapter) Content(ctx context.Context) (any, error) {
	r.init(false, true)
	if r.transformationsErr != nil {
		return nil, r.transformationsErr
	}
	return r.target.Content(ctx)
}

func (r *resourceAdapter) Err() resources.ResourceError {
	return nil
}

func (r *resourceAdapter) GetIdentity() identity.Identity {
	return identity.FirstIdentity(r.target)
}

func (r *resourceAdapter) Data() any {
	r.init(false, false)
	return r.target.Data()
}

func (r *resourceAdapter) ForEeachIdentityByName(name string, f func(identity.Identity) bool) {
	if constants.IsFieldRelOrPermalink(name) && !r.resourceTransformations.hasTransformationPermalinkHash() {
		// Special case for links without any content hash in the URL.
		// We don't need to rebuild all pages that use this resource,
		// but we want to make sure that the resource is accessed at least once.
		f(identity.NewFindFirstManagerIdentityProvider(r.target.GetDependencyManager(), r.target.GetIdentityGroup()))
		return
	}
	f(r.target.GetIdentityGroup())
	f(r.target.GetDependencyManager())
}

func (r *resourceAdapter) GetIdentityGroup() identity.Identity {
	return r.target.GetIdentityGroup()
}

func (r *resourceAdapter) GetDependencyManager() identity.Manager {
	return r.target.GetDependencyManager()
}

func (r *resourceAdapter) cloneTo(targetPath string) resources.Resource {
	newtTarget := r.target.CloneTo(targetPath)
	newInner := &ResourceAdapterInner{
		ctx:    r.ctx,
		Staler: r.Staler,
		target: newtTarget.(transformableResource),
	}
	if r.ResourceAdapterInner.PublishOnce != nil {
		newInner.PublishOnce = &valueobject.PublishOnce{}
	}
	r.ResourceAdapterInner = newInner
	return r
}

func (r *resourceAdapter) Process(spec string) (resources.ImageResource, error) {
	return r.getImageOps().Process(spec)
}

func (r *resourceAdapter) Crop(spec string) (resources.ImageResource, error) {
	return r.getImageOps().Crop(spec)
}

func (r *resourceAdapter) Fill(spec string) (resources.ImageResource, error) {
	return r.getImageOps().Fill(spec)
}

func (r *resourceAdapter) Fit(spec string) (resources.ImageResource, error) {
	return r.getImageOps().Fit(spec)
}

func (r *resourceAdapter) Filter(filters ...any) (resources.ImageResource, error) {
	return r.getImageOps().Filter(filters...)
}

func (r *resourceAdapter) Height() int {
	return r.getImageOps().Height()
}

func (r *resourceAdapter) Exif() *exif.ExifInfo {
	return r.getImageOps().Exif()
}

func (r *resourceAdapter) Colors() ([]string, error) {
	return r.getImageOps().Colors()
}

func (r *resourceAdapter) Key() string {
	r.init(false, false)
	return r.target.(resources.Identifier).Key()
}

func (r *resourceAdapter) MediaType() media.Type {
	r.init(false, false)
	return r.target.MediaType()
}

func (r *resourceAdapter) Name() string {
	r.init(false, false)
	return r.metaProvider.Name()
}

func (r *resourceAdapter) NameNormalized() string {
	r.init(false, false)
	return r.target.(resources.NameNormalizedProvider).NameNormalized()
}

func (r *resourceAdapter) Params() maps.Params {
	r.init(false, false)
	return r.metaProvider.Params()
}

func (r *resourceAdapter) Permalink() string {
	r.init(true, false)
	return r.target.Permalink()
}

func (r *resourceAdapter) Publish() error {
	r.init(false, false)

	return r.target.Publish()
}

func (r *resourceAdapter) ReadSeekCloser() (pio.ReadSeekCloser, error) {
	r.init(false, false)
	return r.target.ReadSeekCloser()
}

func (r *resourceAdapter) RelPermalink() string {
	r.init(true, false)
	return r.target.RelPermalink()
}

func (r *resourceAdapter) Resize(spec string) (resources.ImageResource, error) {
	return r.getImageOps().Resize(spec)
}

func (r *resourceAdapter) ResourceType() string {
	r.init(false, false)
	return r.target.ResourceType()
}

func (r *resourceAdapter) String() string {
	return r.Name()
}

func (r *resourceAdapter) Title() string {
	r.init(false, false)
	return r.metaProvider.Title()
}

func (r *resourceAdapter) Transform(t ...valueobject.ResourceTransformation) (valueobject.ResourceTransformer, error) {
	return r.TransformWithContext(context.Background(), t...)
}

func (r *resourceAdapter) TransformWithContext(ctx context.Context, t ...valueobject.ResourceTransformation) (valueobject.ResourceTransformer, error) {
	r.resourceTransformations = &resourceTransformations{
		transformations: append(r.transformations, t...),
	}

	r.ResourceAdapterInner = &ResourceAdapterInner{
		ctx:         ctx,
		Staler:      r.Staler,
		PublishOnce: &valueobject.PublishOnce{},
		target:      r.target,
	}

	return r, nil
}

func (r *resourceAdapter) Width() int {
	return r.getImageOps().Width()
}

func (r *resourceAdapter) DecodeImage() (image.Image, error) {
	return r.getImageOps().DecodeImage()
}

func (r *resourceAdapter) WithResourceMeta(mp resources.MetaProvider) resources.Resource {
	r.metaProvider = mp
	return r
}

func (r *resourceAdapter) getImageOps() resources.ImageResourceOps {
	img, ok := r.target.(resources.ImageResourceOps)
	if !ok {
		if r.MediaType().SubType == "svg" {
			panic("this method is only available for raster images. To determine if an images is SVG, you can do {{ if eq .MediaType.SubType \"svg\" }}{{ end }}")
		}
		fmt.Println(r.MediaType().SubType)
		panic("this method is only available for images resources")
	}
	r.init(false, false)
	return img
}

func (r *resourceAdapter) publish() {
	if r.PublishOnce == nil {
		return
	}

	r.PublisherInit.Do(func() {
		r.PublisherErr = r.target.Publish()

		if r.PublisherErr != nil {
			_ = fmt.Errorf("failed to publish Resource: %s", r.PublisherErr)
		}
	})
}

func (r *resourceAdapter) TransformationKey() string {
	var key string
	for _, tr := range r.transformations {
		key = key + "_" + tr.Key().Value()
	}
	return r.ResourceCache.CleanKey(r.target.Key()) + "_" + helpers.MD5String(key)
}

func (r *resourceAdapter) getOrTransform(publish, setContent bool) error {
	key := r.TransformationKey()
	res, err := r.ResourceCache.CacheResourceTransformation.GetOrCreate(
		key, func(string) (*ResourceAdapterInner, error) {
			return r.transform(key, publish, setContent)
		})
	if err != nil {
		return err
	}

	r.ResourceAdapterInner = res
	return nil
}

func (r *resourceAdapter) transform(key string, publish, setContent bool) (*ResourceAdapterInner, error) {
	cache := r.ResourceCache

	b1 := bp.GetBuffer()
	b2 := bp.GetBuffer()
	defer bp.PutBuffer(b1)
	defer bp.PutBuffer(b2)

	tctx := &valueobject.ResourceTransformationCtx{
		Ctx:                   r.ctx,
		Data:                  make(map[string]any),
		OpenResourcePublisher: r.target.openPublishFileForWriting,
		DependencyManager:     r.target.GetDependencyManager(),
	}

	tctx.InMediaType = r.target.MediaType()
	tctx.OutMediaType = r.target.MediaType()

	startCtx := *tctx
	updates := &valueobject.TransformationUpdate{StartCtx: startCtx}

	var contentrc pio.ReadSeekCloser

	contentrc, err := valueobject.ContentReadSeekerCloser(r.target)
	if err != nil {
		return nil, err
	}

	defer contentrc.Close()

	tctx.From = contentrc
	tctx.To = b1

	tctx.InPath = r.target.TargetPath()
	tctx.SourcePath = strings.TrimPrefix(tctx.InPath, "/")

	counter := 0
	writeToFileCache := false

	var transformedContentr io.Reader

	for i, tr := range r.transformations {
		if i != 0 {
			tctx.InMediaType = tctx.OutMediaType
		}

		mayBeCachedOnDisk := transformationsToCacheOnDisk[tr.Key().Name]
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

		var tryFileCache bool
		if mayBeCachedOnDisk {
			tryFileCache = true
		} else {
			err = tr.Transform(tctx)
			if err != nil && !errors.Is(err, herrors.ErrFeatureNotAvailable) {
				return nil, newErr(err)
			}

			if mayBeCachedOnDisk {
				tryFileCache = true
			}
			if err != nil && !tryFileCache {
				return nil, newErr(err)
			}
		}

		if tryFileCache {
			f := r.target.tryTransformedFileCache(key, updates)
			if f == nil {
				if err != nil {
					return nil, newErr(err)
				}
				return nil, newErr(fmt.Errorf("resource %q not found in file cache", key))
			}
			transformedContentr = f
			updates.SourceFs = cache.FileCache.Fs
			defer f.Close()

			// The reader above is all we need.
			break
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

	if publish {
		publicw, err := r.target.openPublishFileForWriting(updates.TargetPath)
		if err != nil {
			return nil, err
		}
		publishwriters = append(publishwriters, publicw)
	}

	if transformedContentr == nil {
		if writeToFileCache {
			// Also write it to the cache
			fi, metaw, err := cache.WriteMeta(key, updates.ToTransformedResourceMetadata())
			if err != nil {
				return nil, err
			}
			updates.SourceFilename = &fi.Name
			updates.SourceFs = cache.FileCache.Fs
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

	setContent = setContent || !writeToFileCache

	if setContent {
		contentmemw = bp.GetBuffer()
		defer bp.PutBuffer(contentmemw)
		publishwriters = append(publishwriters, pio.ToWriteCloser(contentmemw))
	}

	publishw := pio.NewMultiWriteCloser(publishwriters...)
	_, err = io.Copy(publishw, transformedContentr)
	if err != nil {
		return nil, err
	}
	publishw.Close()

	if setContent {
		s := contentmemw.String()
		updates.Content = &s
	}

	newTarget, err := r.target.cloneWithUpdates(updates)
	if err != nil {
		return nil, err
	}
	r.target = newTarget

	return r.ResourceAdapterInner, nil
}

func (r *resourceAdapter) init(publish, setContent bool) {
	r.initTransform(publish, setContent)
}

func (r *resourceAdapter) initTransform(publish, setContent bool) {
	r.transformationsInit.Do(func() {
		if len(r.transformations) == 0 {
			// Nothing to do.
			return
		}

		if publish {
			// The transformation will write the content directly to
			// the destination.
			r.PublishOnce = nil
		}

		r.transformationsErr = r.getOrTransform(publish, setContent)
		if r.transformationsErr != nil {
			_ = fmt.Errorf("transformation failed: %s", r.transformationsErr)
		}
	})

	if publish && r.PublishOnce != nil {
		r.publish()
	}
}
