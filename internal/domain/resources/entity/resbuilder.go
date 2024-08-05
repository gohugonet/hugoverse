package entity

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"mime"
	"os"
	"path"
)

type resourceBuilder struct {
	relPathname string
	resPaths    valueobject.ResourcePaths

	openReadSeekCloser io.OpenReadSeekCloser
	mediaType          media.Type

	publisher *Publisher

	mediaSvc  resources.MediaTypesConfig
	cache     *Cache
	imageSvc  resources.ImageConfig
	imageProc *valueobject.ImageProcessor
}

func newResourceBuilder(relPathname string, openReadSeekCloser io.OpenReadSeekCloser) *resourceBuilder {
	return &resourceBuilder{
		relPathname:        relPathname,
		openReadSeekCloser: openReadSeekCloser,
	}
}

func (rs *resourceBuilder) withPublisher(publisher *Publisher) *resourceBuilder {
	rs.publisher = publisher
	return rs
}

func (rs *resourceBuilder) withImageService(imageSvc resources.ImageConfig) *resourceBuilder {
	rs.imageSvc = imageSvc
	return rs
}

func (rs *resourceBuilder) withImageProcessor(imageProc *valueobject.ImageProcessor) *resourceBuilder {
	rs.imageProc = imageProc
	return rs
}

func (rs *resourceBuilder) withCache(cache *Cache) *resourceBuilder {
	rs.cache = cache
	return rs
}

func (rs *resourceBuilder) withMediaService(mediaSvc resources.MediaTypesConfig) *resourceBuilder {
	rs.mediaSvc = mediaSvc
	return rs
}

func (rs *resourceBuilder) build() (resources.Resource, error) {
	if rs.openReadSeekCloser == nil {
		return nil, errors.New("OpenReadSeekCloser is nil")
	}

	if rs.relPathname == "" {
		return nil, errors.New("RelPath is empty")
	}

	if err := rs.buildResPaths(); err != nil {
		return nil, err
	}

	if err := rs.buildMediaType(); err != nil {
		return nil, err
	}

	return rs.buildResource()
}

func (rs *resourceBuilder) buildResPaths() error {
	rs.relPathname = paths.ToSlashPreserveLeading(rs.relPathname)

	dir, name := path.Split(rs.relPathname)
	dir = paths.ToSlashPreserveLeading(dir)
	if dir == "/" {
		dir = ""
	}

	rs.resPaths = valueobject.ResourcePaths{
		Dir:           dir,
		BaseDirLink:   "",
		BaseDirTarget: "",

		File: name,
	}

	return nil
}

func (rs *resourceBuilder) buildMediaType() error {
	resPath := paths.Parse("", rs.relPathname)
	ext := resPath.Ext()

	var (
		found      bool
		suffixInfo media.SuffixInfo
	)

	mediaType, suffixInfo, found := rs.mediaSvc.LookFirstBySuffix(ext)
	// TODO(bep) we need to handle these ambiguous types better, but in this context
	// we most likely want the application/xml type.
	if suffixInfo.Suffix == "xml" && mediaType.SubType == "rss" {
		mediaType, found = rs.mediaSvc.LookByType("application/xml")
	}

	if !found {
		// A fallback. Note that mime.TypeByExtension is slow by Hugo standards,
		// so we should configure media types to avoid this lookup for most
		// situations.
		mimeStr := mime.TypeByExtension("." + ext)
		if mimeStr != "" {
			mediaType, _ = media.FromStringAndExt(mimeStr, ext)
		}
	}

	rs.mediaType = mediaType
	return nil
}

func (rs *resourceBuilder) buildResource() (resources.Resource, error) {
	gr := &Resource{
		Staler: &stale.AtomicStaler{},
		h:      &valueobject.ResourceHash{},

		mediaType:          rs.mediaType,
		openReadSeekCloser: rs.openReadSeekCloser,

		paths: rs.resPaths,

		data:              make(map[string]any),
		dependencyManager: identity.NewManager("resource"),
	}

	switch valueobject.ClassifyType(rs.mediaType.Type) {
	case "transformer":
		rt := &ResourceTransformer{
			Resource:  *gr,
			publisher: rs.publisher,

			resourceTransformations: &resourceTransformations{},

			TransformationCache: rs.cache,
		}

		return rt, nil
	case "image":
		imgFormat, ok := valueobject.ImageFormatFromMediaSubType(rs.mediaType.SubType)
		if !ok {
			return nil, &os.PathError{Op: "newResource", Path: rs.relPathname, Err: errors.New("unknown image format")}
		}
		img := valueobject.NewImage(imgFormat, nil, gr)

		ri := &ResourceImage{
			Resource: *gr,
			Image:    img,

			ImageCache:  rs.cache,
			ImageConfig: rs.imageSvc,
			ImageProc:   rs.imageProc,
		}
		ri.root = ri

		return ri, nil
	default:
		return gr, nil
	}
}
