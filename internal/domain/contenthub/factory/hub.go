package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/entity"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
)

func New(tp contenthub.ThemeProvider) (*entity.ContentHub, error) {
	ch := &entity.ContentHub{
		ThemeProvider: tp,
	}

	modules, err := loadModules(tp.Name())
	if err != nil {
		return nil, err
	}
	ch.Modules = modules

	return ch, nil
}

func loadModules(theme string) (module.Modules, error) {
	return mdFact.New(theme)
}
