package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

func NewWalkway(fs afero.Fs, root string, walker valueobject.WalkFunc) *valueobject.Walkway {
	return &valueobject.Walkway{
		Fs:     fs,
		Root:   root,
		WalkFn: walker,
		Seen:   make(map[string]bool),

		Log: log,
	}
}
