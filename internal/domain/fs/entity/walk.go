package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

func (f *Fs) WalkAssets(start string, cb fs.WalkCallback, conf valueobject.WalkwayConfig) error {
	return f.Walk(f.Assets, start, cb, conf)
}

func (f *Fs) Walk(fs afero.Fs, start string, cb fs.WalkCallback, conf valueobject.WalkwayConfig) error {
	w, err := valueobject.NewWalkway(fs, cb)
	if err != nil {
		return err
	}

	return w.WalkWith(start, conf)
}
