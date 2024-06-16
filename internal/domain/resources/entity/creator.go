package entity

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"os"
	"path"
)

func (rs *Resources) newResource(rd *valueobject.ResourceSourceDescriptor) (resources.Resource, error) {
	dir, name := path.Split(rd.TargetPath)
	dir = paths.ToSlashPreserveLeading(dir)
	if dir == "/" {
		dir = ""
	}

	rp := valueobject.ResourcePaths{
		File:                    name,
		Dir:                     dir,
		BaseDirTarget:           rd.BasePathTargetPath,
		BaseDirLink:             rd.BasePathRelPermalink,
		BasePathNoTrailingSlash: rs.UrlService.BasePathNoSlash(),
		TargetBasePaths:         rd.TargetBasePaths,
	}

	gr := &Resource{
		Staler: &stale.AtomicStaler{},
		h:      &valueobject.ResourceHash{},

		paths:  rp,
		sd:     *rd,
		params: make(map[string]any),
		name:   rd.NameOriginal,
		title:  rd.NameOriginal,

		publishFs: rs.FsService.PublishFs(),
	}

	switch valueobject.ClassifyType(rd.MediaType.Type) {
	case "transformer":
		rt := &ResourceTransformer{
			Resource:                *gr,
			resourceTransformations: &resourceTransformations{},

			TransformationCache: rs.Cache,
		}

		return rt, nil
	case "image":
		imgFormat, ok := valueobject.ImageFormatFromMediaSubType(rd.MediaType.SubType)
		if !ok {
			return nil, &os.PathError{Op: "newResource", Path: rd.TargetPath, Err: errors.New("unknown image format")}
		}
		img := valueobject.NewImage(imgFormat, nil, gr)

		ri := &ResourceImage{
			Resource: *gr,
			Image:    img,

			ImageCache:  rs.Cache,
			ImageConfig: rs.ImageService,
			ImageProc:   rs.ImageProc,
		}
		ri.root = ri

		return ri, nil
	default:
		return gr, nil
	}
}
