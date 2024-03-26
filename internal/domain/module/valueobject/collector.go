package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/log"
)

type ModuleCollector struct {
	Modules []module.Module
	log     log.Logger
}

func NewModuleCollector(log log.Logger) *ModuleCollector {
	return &ModuleCollector{
		Modules: []module.Module{},
		log:     log,
	}
}

func (mc *ModuleCollector) CollectModules(modConfig module.ModuleConfig, hookBeforeFinalize func(m []module.Module)) {
	projectMod := &moduleAdapter{
		projectMod: true,
		config:     modConfig,
	}

	// module structure, [project, others...]
	mc.addAndRecurse(projectMod)

	// Add the project mod on top.
	mc.Modules = append([]module.Module{projectMod}, mc.Modules...)

	if hookBeforeFinalize != nil {
		hookBeforeFinalize(mc.Modules)
	}
}

// addAndRecurse Project Imports -> Import imports
func (mc *ModuleCollector) addAndRecurse(owner *moduleAdapter) {
	moduleConfig := owner.Config()

	// theme may depend on other theme
	for _, moduleImport := range moduleConfig.Imports {
		tc := mc.add(owner, moduleImport)
		if tc == nil {
			continue
		}
		// tc is mytheme with no config file
		mc.addAndRecurse(tc)
	}
}

func (mc *ModuleCollector) add(owner *moduleAdapter, moduleImport module.Import) *moduleAdapter {
	mc.log.Printf("--- start to create `%s` module", moduleImport.Path)

	ma := &moduleAdapter{
		owner: owner,
		// In the example, "mytheme" has no other import
		// In the real world, we need to parse the theme config and download the theme repo
		config: module.ModuleConfig{},
	}
	mc.Modules = append(mc.Modules, ma)
	return ma
}
