package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/glob"
	"github.com/gohugonet/hugoverse/pkg/hexec"
	"github.com/gohugonet/hugoverse/pkg/io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Resources struct {
	*Cache
	*Publisher

	ExecHelper *hexec.Exec

	*Common

	FsService    resources.Fs
	MediaService resources.MediaTypesConfig

	ImageService resources.ImageConfig
	ImageProc    *valueobject.ImageProcessor

	*MinifierClient
	*TemplateClient
	*IntegrityClient
	*SassClient
}

func (rs *Resources) SetupTemplateClient(tmpl Template) {
	rs.TemplateClient = &TemplateClient{tmpl}
}

func (rs *Resources) GetResourceWithOpener(pathname string, opener io.OpenReadSeekCloser) (resources.Resource, error) {
	pathname = path.Clean(pathname)
	key := dynacache.CleanKey(pathname) + "__get"

	return rs.Cache.GetOrCreateResource(key, func() (resources.Resource, error) {
		rsb := newResourceBuilder(pathname, opener)
		rsb.withCache(rs.Cache).withMediaService(rs.MediaService).
			withImageService(rs.ImageService).withImageProcessor(rs.ImageProc).
			withPublisher(rs.Publisher)

		return rsb.build()
	})

	// TODO: analysis the impact of default assets filesystem
	// Will impact the css types source processing in the future
}

func (rs *Resources) GetResource(pathname string) (resources.Resource, error) {
	pathname = path.Clean(pathname)
	key := dynacache.CleanKey(pathname) + "__get"

	return rs.Cache.GetOrCreateResource(key, func() (resources.Resource, error) {
		// The resource file will not be read before it gets used (e.g. in .Content),
		// so we need to check that the file exists here.
		filename := filepath.FromSlash(pathname)
		_, err := rs.FsService.AssetsFs().Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			// A real error.
			return nil, err
		}

		rsb := newResourceBuilder(pathname, func() (io.ReadSeekCloser, error) {
			return rs.FsService.AssetsFs().Open(filename)
		})
		rsb.withCache(rs.Cache).withMediaService(rs.MediaService).
			withImageService(rs.ImageService).withImageProcessor(rs.ImageProc).
			withPublisher(rs.Publisher)

		return rsb.build()
	})
}

// GetMatch gets first resource matching the given pattern from the assets filesystem.
func (rs *Resources) GetMatch(pattern string) (resources.Resource, error) {
	res, err := rs.match("__get-match", pattern, nil, true)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	return res[0], err
}

func (rs *Resources) match(name, pattern string, matchFunc func(r resources.Resource) bool, firstOnly bool) ([]resources.Resource, error) {
	pattern = glob.NormalizePath(pattern)
	partitions := glob.FilterGlobParts(strings.Split(pattern, "/"))
	key := path.Join(name, path.Join(partitions...))
	key = path.Join(key, pattern)

	return rs.Cache.GetOrCreateResources(key, func() ([]resources.Resource, error) {
		var res []resources.Resource

		handle := func(info fs.FileMetaInfo) (bool, error) {
			rsb := newResourceBuilder(info.FileName(), func() (io.ReadSeekCloser, error) {
				return rs.FsService.AssetsFs().Open(info.FileName())
			})
			rsb.withCache(rs.Cache).withMediaService(rs.MediaService).
				withImageService(rs.ImageService).withImageProcessor(rs.ImageProc).
				withPublisher(rs.Publisher)

			r, err := rsb.build()
			if err != nil {
				return true, err
			}

			if matchFunc != nil && !matchFunc(r) {
				return false, nil
			}

			res = append(res, r)

			return firstOnly, nil
		}

		if err := rs.FsService.Glob(rs.FsService.AssetsFs(), pattern, handle); err != nil {
			return nil, err
		}

		return res, nil
	})
}

// Copy copies r to the new targetPath.
func (rs *Resources) Copy(r resources.Resource, targetPath string) (resources.Resource, error) {
	key := dynacache.CleanKey(targetPath) + "__copy"
	return rs.Cache.GetOrCreateResource(key, func() (resources.Resource, error) {
		return r.(resources.ResourceCopier).CloneTo(targetPath), nil
	})
}
