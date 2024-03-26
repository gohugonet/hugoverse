package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/internal/domain/module/entity"
)

func New(theme string) (module.Modules, error) {
	ms := &entity.Module{Theme: theme}
	ms.SetupLog()

	if err := ms.Load(); err != nil {
		return nil, err
	}
	return ms, nil
}
