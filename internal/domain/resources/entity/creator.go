package entity

import (
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"go.opencensus.io/resource"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Creator struct {
	MediaService resources.MediaTypes
	UrlService   resources.Url
	AssetsFs     afero.Fs
	PublishFs    afero.Fs

	HttpClient       *http.Client
	CacheGetResource *filecache.Cache
	ResourceCache    *valueobject.ResourceCache
}

// Copy copies r to the new targetPath.
func (c *Creator) Copy(r resources.Resource, targetPath string) (resources.Resource, error) {
	key := dynacache.CleanKey(targetPath) + "__copy"
	return c.ResourceCache.GetOrCreate(key, func() (resources.Resource, error) {
		return valueobject.Copy(r, targetPath), nil
	})
}

// Get creates a new Resource by opening the given pathname in the assets filesystem.
func (c *Creator) Get(pathname string) (resources.Resource, error) {
	pathname = path.Clean(pathname)
	key := dynacache.CleanKey(pathname) + "__get"

	return c.ResourceCache.GetOrCreate(key, func() (resources.Resource, error) {
		// The resource file will not be read before it gets used (e.g. in .Content),
		// so we need to check that the file exists here.
		filename := filepath.FromSlash(pathname)
		fi, err := c.AssetsFs.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			// A real error.
			return nil, err
		}

		// TODO, refactor PathInfo
		pi := fi.(fsVO.FileMetaInfo).Meta().PathInfo

		return c.newResource(valueobject.ResourceSourceDescriptor{
			MediaService: c.MediaService,
			LazyPublish:  true,
			OpenReadSeekCloser: func() (io.ReadSeekCloser, error) {
				return c.AssetsFs.Open(filename)
			},
			Path:          pi,
			GroupIdentity: pi,
			TargetPath:    pathname,
		})
	})
}

func (c *Creator) newResource(rd valueobject.ResourceSourceDescriptor) (resources.Resource, error) {
	if err := rd.Setup(); err != nil {
		return nil, err
	}

	dir, name := path.Split(rd.TargetPath)
	dir = paths.ToSlashPreserveLeading(dir)
	if dir == "/" {
		dir = ""
	}

	rp := valueobject.ResourcePaths{
		File:            name,
		Dir:             dir,
		BaseDirTarget:   rd.BasePathTargetPath,
		BaseDirLink:     rd.BasePathRelPermalink,
		TargetBasePaths: rd.TargetBasePaths,
	}

	gr := &genericResource{
		UrlService:   c.UrlService,
		MediaService: c.MediaService,

		Staler: &valueobject.AtomicStaler{},
		h:      &valueobject.ResourceHash{},

		resourceCache: c.ResourceCache,

		publishFs:   c.PublishFs,
		publishInit: &sync.Once{},
		paths:       rp,
		sd:          rd,
		params:      make(map[string]any),
		name:        rd.NameOriginal,
		title:       rd.NameOriginal,
	}

	if rd.MediaType.MainType == "image" {
		imgFormat, ok := images.ImageFormatFromMediaSubType(rd.MediaType.SubType)
		if ok {
			ir := &imageResource{
				Image:        images.NewImage(imgFormat, r.imaging, nil, gr),
				baseResource: gr,
			}
			ir.root = ir
			return newResourceAdapter(gr.spec, rd.LazyPublish, ir), nil
		}

	}

	return newResourceAdapter(gr.spec, rd.LazyPublish, gr), nil
}

// Match gets the resources matching the given pattern from the assets filesystem.
func (c *Creator) Match(pattern string) (resource.Resources, error) {
	return c.match("__match", pattern, nil, false)
}

func (c *Client) ByType(tp string) resource.Resources {
	res, err := c.match(path.Join("_byType", tp), "**", func(r resource.Resource) bool { return r.ResourceType() == tp }, false)
	if err != nil {
		panic(err)
	}
	return res
}

// GetMatch gets first resource matching the given pattern from the assets filesystem.
func (c *Creator) GetMatch(pattern string) (resource.Resource, error) {
	res, err := c.match("__get-match", pattern, nil, true)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	return res[0], err
}

func (c *Creator) match(name, pattern string, matchFunc func(r resource.Resource) bool, firstOnly bool) (resource.Resources, error) {
	pattern = glob.NormalizePath(pattern)
	partitions := glob.FilterGlobParts(strings.Split(pattern, "/"))
	key := path.Join(name, path.Join(partitions...))
	key = path.Join(key, pattern)

	return c.rs.ResourceCache.GetOrCreateResources(key, func() (resource.Resources, error) {
		var res resource.Resources

		handle := func(info hugofs.FileMetaInfo) (bool, error) {
			meta := info.Meta()

			r, err := c.rs.NewResource(resources.ResourceSourceDescriptor{
				LazyPublish: true,
				OpenReadSeekCloser: func() (hugio.ReadSeekCloser, error) {
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

		if err := hugofs.Glob(c.rs.BaseFs.Assets.Fs, pattern, handle); err != nil {
			return nil, err
		}

		return res, nil
	})
}

// FromString creates a new Resource from a string with the given relative target path.
// TODO(bep) see #10912; we currently emit a warning for this config scenario.
func (c *Creator) FromString(targetPath, content string) (resource.Resource, error) {
	targetPath = path.Clean(targetPath)
	key := dynacache.CleanKey(targetPath) + helpers.MD5String(content)
	r, err := c.rs.ResourceCache.GetOrCreate(key, func() (resource.Resource, error) {
		return c.rs.NewResource(
			resources.ResourceSourceDescriptor{
				LazyPublish:   true,
				GroupIdentity: identity.Anonymous, // All usage of this resource are tracked via its string content.
				OpenReadSeekCloser: func() (hugio.ReadSeekCloser, error) {
					return hugio.NewReadSeekerNoOpCloserFromString(content), nil
				},
				TargetPath: targetPath,
			})
	})

	return r, err
}
