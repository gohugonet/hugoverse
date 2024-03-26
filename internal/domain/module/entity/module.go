package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/internal/domain/module/valueobject"
	"github.com/gohugonet/hugoverse/pkg/log"
)

type Module struct {
	Theme   string
	modules []module.Module
	log     log.Logger
}

func (m *Module) Proj() module.Module {
	return m.modules[0]
}

func (m *Module) All() []module.Module {
	return m.modules
}

func (m *Module) SetupLog() {
	m.log = log.NewStdLogger()
}

func (m *Module) Load() error {
	if m.Theme != "" {
		ms, err := m.loadModules(m.Theme)
		if err != nil {
			return err
		}
		m.modules = ms
		return nil
	}

	m.log.Errorf("empty theme")
	return fmt.Errorf("empty theme")
}

func (m *Module) loadModules(theme string) ([]module.Module, error) {
	// project module config
	projModuleConfig := module.ModuleConfig{}
	imports := []string{theme}
	for _, imp := range imports {
		projModuleConfig.Imports = append(
			projModuleConfig.Imports, module.Import{
				Path: imp,
			})
	}

	mc := valueobject.NewModuleCollector(m.log)
	// Need to run these after the modules are loaded, but before
	// they are finalized.
	collectHook := func(mods []module.Module) {
		// Apply default project mounts.
		// Default folder structure for hugo project
		for i, mod := range mods {
			if mod.IsProj() {
				valueobject.ApplyProjectConfigDefaults(mod)
			}
			m.log.Printf("Apply default project mounts: %d, %+v", i, mod)
		}
	}
	mc.CollectModules(projModuleConfig, collectHook)

	return mc.Modules, nil
}
