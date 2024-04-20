package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
)

type Module struct {
	valueobject.ModuleConfig
}

func (m Module) ImportPaths() []string {
	var paths []string
	for _, i := range m.ModuleConfig.Imports {
		paths = append(paths, i.Path)
	}
	return paths
}
