package valueobject

import (
	"github.com/spf13/afero"
	"os"
)

func NewBaseFs(fs afero.Fs) afero.Fs {
	return &baseFs{Fs: fs}
}

type baseFs struct {
	afero.Fs
}

func (fs *baseFs) UnwrapFilesystem() afero.Fs {
	return fs.Fs
}

func (fs *baseFs) Stat(name string) (os.FileInfo, error) {
	fi, err := fs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}

	fim := NewFileInfoWithOpener(fi, name, func() (afero.File, error) {
		if fi.IsDir() {
			return fs.openDir(name)
		}
		return fs.open(name)
	})

	return fim, nil
}

func (fs *baseFs) Open(name string) (afero.File, error) {
	return fs.open(name)
}

func (fs *baseFs) open(name string) (afero.File, error) {
	f, err := fs.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return NewFile(f), nil
}

func (fs *baseFs) openDir(name string) (afero.File, error) {
	f, err := fs.open(name)
	if err != nil {
		return nil, err
	}
	return NewDirFile(f), nil
}
