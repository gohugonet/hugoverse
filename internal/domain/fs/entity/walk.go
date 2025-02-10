package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

func (f *Fs) WalkAssets(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error {
	return f.Walk(f.Assets, start, cb, conf)
}

func (f *Fs) WalkContent(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error {
	return f.Walk(f.Content, start, cb, conf)
}

func (f *Fs) WalkLayouts(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error {
	return f.Walk(f.Layouts, start, cb, conf)
}

func (f *Fs) WalkI18n(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error {
	return f.Walk(f.I18n, start, cb, conf)
}

func (f *Fs) Walk(fs afero.Fs, start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error {
	w, err := valueobject.NewWalkway(fs, cb)
	if err != nil {
		return err
	}

	return w.WalkWith(start, conf)
}
