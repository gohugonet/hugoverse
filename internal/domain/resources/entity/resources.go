package entity

import (
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/hexec"
	"github.com/gohugonet/hugoverse/pkg/io"
	"os"
	"path"
	"path/filepath"
)

type Resources struct {
	*Cache

	ExecHelper *hexec.Exec

	*Common

	FsService    resources.Fs
	MediaService resources.MediaTypes
	UrlService   resources.Url
	GlobService  resources.Glob

	ImageService resources.ImageConfig
	ImageProc    *valueobject.ImageProcessor
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
		pi := fi.(fsVO.FileMetaInfo).Meta().PathInfo

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
