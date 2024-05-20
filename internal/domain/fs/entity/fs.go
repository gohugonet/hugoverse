package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

type Fs struct {
	*OriginFs
	*PathSpec

	Content    *valueobject.ComponentFs
	Data       *valueobject.ComponentFs
	I18n       *valueobject.ComponentFs
	Layouts    *valueobject.ComponentFs
	Archetypes *valueobject.ComponentFs
	Assets     *valueobject.ComponentFs
}

func (f *Fs) NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return valueobject.NewBasePathFs(source, path)
}

func (f *Fs) Glob(fs afero.Fs, pattern string, handle func(fi valueobject.FileMetaInfo) (bool, error)) error {
	return valueobject.Glob(fs, pattern, handle)
}
