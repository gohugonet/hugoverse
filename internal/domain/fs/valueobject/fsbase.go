package valueobject

import (
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"os"
)

func NewBaseFs(fs afero.Fs) afero.Fs {
	return &baseFs{Fs: fs, log: loggers.NewDefault()}
}

type baseFs struct {
	afero.Fs // osFs

	log loggers.Logger
}

func (fs *baseFs) UnwrapFilesystem() afero.Fs {
	return fs.Fs
}

func (fs *baseFs) Stat(absName string) (os.FileInfo, error) {
	fi, err := fs.Fs.Stat(absName)
	if err != nil {
		return nil, err
	}

	var ofi os.FileInfo

	if fi.IsDir() {
		ofi = NewFileInfoWithOpener(fi, absName, func() (afero.File, error) {
			return fs.openDir(absName)
		})

		return ofi, nil
	}

	ofi = NewFileInfoWithOpener(fi, absName, func() (afero.File, error) {
		return fs.open(absName)
	})

	return ofi, nil
}

func (fs *baseFs) Open(name string) (afero.File, error) {
	return fs.open(name)
}

func (fs *baseFs) open(name string) (*File, error) {
	f, err := fs.Fs.Open(name)

	if err != nil {
		return nil, err
	}

	return NewFile(f, name), nil
}

func (fs *baseFs) openDir(name string) (afero.File, error) {
	f, err := fs.open(name)
	if err != nil {
		return nil, err
	}

	return NewDirFile(f, FileMeta{filename: name}, fs), nil
}
