package entity

import (
	"errors"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/glob"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"net/http"
	"os"
	"path"
	"strings"
)

type Creator struct {
	MediaService resources.MediaTypes
	UrlService   resources.Url
	GlobService  resources.Glob

	AssetsFs  afero.Fs
	PublishFs afero.Fs

	HttpClient *http.Client

	Imaging *valueobject.ImageProcessor
}

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
		Staler: &valueobject.AtomicStaler{},
		h:      &valueobject.ResourceHash{},

		paths:  rp,
		sd:     *rd,
		params: make(map[string]any),
		name:   rd.NameOriginal,
		title:  rd.NameOriginal,
	}

	switch valueobject.ClassifyType(rd.MediaType.Type) {
	case "transformer":
		rt := ResourceTransformer{
			Resource:                *gr,
			resourceTransformations: &resourceTransformations{},
		}
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

	return newResourceAdapter(c.ResourceCache, gr.spec, rd.LazyPublish, gr), nil
}

// Match gets the resources matching the given pattern from the assets filesystem.
func (c *Creator) Match(pattern string) ([]resources.Resource, error) {
	return c.match("__match", pattern, nil, false)
}

func (c *Creator) ByType(tp string) []resources.Resource {
	res, err := c.match(path.Join("_byType", tp), "**", func(r resources.Resource) bool { return r.ResourceType() == tp }, false)
	if err != nil {
		panic(err)
	}
	return res
}

// GetMatch gets first resource matching the given pattern from the assets filesystem.
func (c *Creator) GetMatch(pattern string) (resources.Resource, error) {
	res, err := c.match("__get-match", pattern, nil, true)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	return res[0], err
}

func (c *Creator) match(name, pattern string, matchFunc func(r resources.Resource) bool, firstOnly bool) ([]resources.Resource, error) {
	pattern = glob.NormalizePath(pattern)
	partitions := glob.FilterGlobParts(strings.Split(pattern, "/"))
	key := path.Join(name, path.Join(partitions...))
	key = path.Join(key, pattern)

	return c.ResourceCache.GetOrCreateResources(key, func() ([]resources.Resource, error) {
		var res []resources.Resource

		handle := func(info fsVO.FileMetaInfo) (bool, error) {
			meta := info.Meta()

			r, err := c.newResource(valueobject.ResourceSourceDescriptor{
				LazyPublish: true,
				OpenReadSeekCloser: func() (io.ReadSeekCloser, error) {
					return meta.Open()
				},
				NameNormalized: meta.PathInfo.Path(),
				NameOriginal:   meta.PathInfo.Unnormalized().Path(),
				GroupIdentity:  meta.PathInfo,
				TargetPath:     meta.PathInfo.Unnormalized().Path(),
			})
			if err != nil {
				return true, err
			}

			if matchFunc != nil && !matchFunc(r) {
				return false, nil
			}

			res = append(res, r)

			return firstOnly, nil
		}

		if err := c.GlobService.Glob(c.AssetsFs, pattern, handle); err != nil {
			return nil, err
		}

		return res, nil
	})
}

// FromString creates a new Resource from a string with the given relative target path.
// TODO(bep) see #10912; we currently emit a warning for this config scenario.
func (c *Creator) FromString(targetPath, content string) (resources.Resource, error) {
	targetPath = path.Clean(targetPath)
	key := dynacache.CleanKey(targetPath) + helpers.MD5String(content)
	r, err := c.ResourceCache.GetOrCreate(key, func() (resources.Resource, error) {
		return c.newResource(
			valueobject.ResourceSourceDescriptor{
				LazyPublish:   true,
				GroupIdentity: identity.Anonymous, // All usage of this resource are tracked via its string content.
				OpenReadSeekCloser: func() (io.ReadSeekCloser, error) {
					return io.NewReadSeekerNoOpCloserFromString(content), nil
				},
				TargetPath: targetPath,
			})
	})

	return r, err
}
