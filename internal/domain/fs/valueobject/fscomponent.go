package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"os"
)

type ComponentFs struct {
	afero.Fs

	// The component name, e.g. "content", "layouts" etc.
	Component string

	OverlayFs afero.Fs

	// The parser used to parse paths provided by this filesystem.
	PathParser *paths.PathParser
}

func (cfs *ComponentFs) UnwrapFilesystem() afero.Fs {
	return cfs.Fs
}

func (cfs *ComponentFs) Stat(name string) (os.FileInfo, error) {
	fi, err := cfs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}
	fim, _ := cfs.applyMeta(fi, name)
	return fim, nil
}

func (cfs *ComponentFs) applyMeta(fi os.FileInfo, name string) (fs.FileMetaInfo, bool) {
	name = normalizeFilename(name)
	fim := fi.(fs.FileMetaInfo)

	if fi.IsDir() {
		opener := func() (afero.File, error) {
			return cfs.Open(name)
		}
		fim = NewFileInfoWithOpener(fi, name, opener)
	}

	return fim, true
}
