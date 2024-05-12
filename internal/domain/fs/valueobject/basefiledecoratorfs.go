package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
	"path/filepath"

	iofs "io/fs"
)

func decorateDirs(fs afero.Fs, meta *FileMeta) afero.Fs {
	ffs := &BaseFileDecoratorFs{Fs: fs}

	decorator := func(fi FileNameIsDir, name string) (FileNameIsDir, error) {
		if !fi.IsDir() {
			// Leave regular files as they are.
			return fi, nil
		}

		return DecorateFileInfo(fi, nil, "", meta), nil
	}

	ffs.Decorate = decorator

	return ffs
}

type BaseFileDecoratorFs struct {
	afero.Fs
	Decorate func(fi FileNameIsDir, name string) (FileNameIsDir, error)
}

func (fs *BaseFileDecoratorFs) UnwrapFilesystem() afero.Fs {
	return fs.Fs
}

func (fs *BaseFileDecoratorFs) Stat(name string) (os.FileInfo, error) {
	fi, err := fs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}

	fim, err := fs.Decorate(fi, name)
	if err != nil {
		return nil, err
	}
	return fim.(os.FileInfo), nil
}

func (fs *BaseFileDecoratorFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	var (
		fi  os.FileInfo
		err error
		ok  bool
	)

	if lstater, isLstater := fs.Fs.(afero.Lstater); isLstater {
		fi, ok, err = lstater.LstatIfPossible(name)
	} else {
		fi, err = fs.Fs.Stat(name)
	}

	if err != nil {
		return nil, false, err
	}

	fim, err := fs.Decorate(fi, name)
	if err != nil {
		return nil, false, err
	}

	return fim.(os.FileInfo), ok, err
}

func (fs *BaseFileDecoratorFs) Open(name string) (afero.File, error) {
	f, err := fs.Fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &baseFileDecoratorFile{File: f, fs: fs}, nil
}

type baseFileDecoratorFile struct {
	afero.File
	fs *BaseFileDecoratorFs
}

func (l *baseFileDecoratorFile) ReadDir(n int) ([]iofs.DirEntry, error) {
	fis, err := l.File.(iofs.ReadDirFile).ReadDir(-1)
	if err != nil {
		return nil, err
	}

	fisp := make([]iofs.DirEntry, len(fis))

	for i, fi := range fis {
		filename := fi.Name()
		if l.Name() != "" {
			filename = filepath.Join(l.Name(), fi.Name())
		}

		fid, err := l.fs.Decorate(fi, filename)
		if err != nil {
			return nil, fmt.Errorf("decorate: %w", err)
		}

		fisp[i] = fid.(iofs.DirEntry)

	}

	return fisp, err
}

func (l *baseFileDecoratorFile) Readdir(c int) (ofi []os.FileInfo, err error) {
	dirEntry, err := l.ReadDir(c)
	if err != nil {
		return nil, err
	}
	var result []os.FileInfo
	for _, d := range dirEntry {
		result = append(result, d.(os.FileInfo))
	}
	return result, nil
}
