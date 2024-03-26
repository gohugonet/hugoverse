package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

func decorateDirs(fs afero.Fs, meta *FileMeta) afero.Fs {
	ffs := &BaseFileDecoratorFs{Fs: fs}

	decorator := func(fi os.FileInfo, name string) (os.FileInfo, error) {
		if !fi.IsDir() {
			// Leave regular files as they are.
			return fi, nil
		}

		return DecorateFileInfo(fi, fs, nil, "", "", meta), nil
	}

	ffs.Decorate = decorator

	return ffs
}

type BaseFileDecoratorFs struct {
	afero.Fs
	Decorate func(fi os.FileInfo, filename string) (os.FileInfo, error)
}

func (fs *BaseFileDecoratorFs) UnwrapFilesystem() afero.Fs {
	return fs.Fs
}

func (fs *BaseFileDecoratorFs) Stat(name string) (os.FileInfo, error) {
	fi, err := fs.Fs.Stat(name)
	if err != nil {
		return nil, err
	}

	return fs.Decorate(fi, name)
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

	fi, err = fs.Decorate(fi, name)

	return fi, ok, err
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

func (l *baseFileDecoratorFile) Readdir(c int) (ofi []os.FileInfo, err error) {
	dirnames, err := l.File.Readdirnames(c)
	if err != nil {
		return nil, err
	}

	fisp := make([]os.FileInfo, 0, len(dirnames))

	for _, dirname := range dirnames {
		filename := dirname

		if l.Name() != "" && l.Name() != fs.FilepathSeparator {
			filename = filepath.Join(l.Name(), dirname)
		}

		// We need to resolve any symlink info.
		fi, _, err := LstatIfPossible(l.fs.Fs, filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		fi, err = l.fs.Decorate(fi, filename)
		if err != nil {
			return nil, fmt.Errorf("Decorate: %w", err)
		}
		fisp = append(fisp, fi)
	}

	return fisp, err
}
