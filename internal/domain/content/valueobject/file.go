package valueobject

import (
	"github.com/spf13/afero"
	"path/filepath"
)

type File struct {
	Fs      afero.Fs
	Path    string
	Content []byte
}

func (f *File) Dump() error {
	err := f.Fs.MkdirAll(filepath.Dir(f.Path), 0777)
	if err != nil {
		return err
	}
	err = afero.WriteFile(f.Fs, f.Path, f.Content, 0666)
	if err != nil {
		return err
	}

	return nil
}
