package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
	"path/filepath"
)

// NewBaseFileDecorator decorates the given Fs to provide the real filename
// and an Opener func.
func NewBaseFileDecorator(originFs afero.Fs) afero.Fs {
	ffs := &valueobject.BaseFileDecoratorFs{Fs: originFs}

	decorator := func(fi valueobject.FileNameIsDir, filename string) (valueobject.FileNameIsDir, error) {
		// Store away the original in case it's a symlink.
		meta := valueobject.NewFileMeta()
		meta.Name = fi.Name()

		if fi.IsDir() {
			meta.JoinStatFunc = func(name string) (valueobject.FileMetaInfo, error) {
				joinedFilename := filepath.Join(filename, name)
				fi, _, err := valueobject.LstatIfPossible(originFs, joinedFilename)
				if err != nil {
					return nil, err
				}

				fim, err := ffs.Decorate(fi, joinedFilename)
				if err != nil {
					return nil, err
				}

				return fim.(valueobject.FileMetaInfo), nil
			}
		}

		opener := func() (afero.File, error) {
			return ffs.Open(filename)
		}

		fim := valueobject.DecorateFileInfo(fi, opener, filename, meta)

		return fim, nil
	}

	ffs.Decorate = decorator
	return ffs
}
