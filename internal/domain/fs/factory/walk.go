package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

func NewWalkway(fs afero.Fs, root string, walker valueobject.WalkFunc) *valueobject.Walkway {
	return valueobject.NewWalkway(fs, root, walker)
}
