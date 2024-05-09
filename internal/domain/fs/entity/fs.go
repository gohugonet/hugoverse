package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

type Fs struct {
	*OriginFs
	*PathSpec
}

func (f *Fs) NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return valueobject.NewBasePathFs(source, path)
}

func (f *Fs) Glob(fs afero.Fs, pattern string, handle func(fi valueobject.FileMetaInfo) (bool, error)) error {
	return valueobject.Glob(fs, pattern, handle)
}
