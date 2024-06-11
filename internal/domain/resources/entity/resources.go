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

	ExecHelper *hexec.Exec

	*Common

	FsService    resources.Fs
	MediaService resources.MediaTypesConfig
	UrlService   resources.SiteUrl

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

func (rs *Resources) GetResource(pathname string) (resources.Resource, error) {
	pathname = path.Clean(pathname)
	key := dynacache.CleanKey(pathname) + "__get"

	return rs.Cache.GetOrCreateResource(key, func() (resources.Resource, error) {
		// The resource file will not be read before it gets used (e.g. in .Content),
		// so we need to check that the file exists here.
		filename := filepath.FromSlash(pathname)
		fi, err := rs.FsService.AssetsFs().Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			// A real error.
			return nil, err
		}

		// TODO, refactor PathInfo
		pi := fi.(fs.FileMetaInfo).Path()

		sd, err := valueobject.NewResourceSourceDescriptor(
			pathname, pi, rs.MediaService,
			func() (io.ReadSeekCloser, error) {
				return rs.FsService.AssetsFs().Open(filename)
			},
		)
		if err != nil {
			return nil, err
		}
		return rs.newResource(sd)
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
			pinfo := info.Path()

			r, err := rs.newResource(&valueobject.ResourceSourceDescriptor{
				LazyPublish: true,
				OpenReadSeekCloser: func() (io.ReadSeekCloser, error) {
					return meta.Open()
				},
				NameNormalized: pinfo.Path(),
				NameOriginal:   pinfo.Unnormalized().Path(),
				GroupIdentity:  pinfo,
				TargetPath:     pinfo.Unnormalized().Path(),
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
		return valueobject.Copy(r, targetPath), nil
	})
}
