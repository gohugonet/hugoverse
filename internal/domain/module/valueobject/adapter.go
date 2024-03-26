package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
)

// moduleAdapter implemented Module interface
type moduleAdapter struct {
	projectMod bool
	owner      module.Module
	mounts     []module.Mount
	config     module.ModuleConfig
}

func (m *moduleAdapter) Config() module.ModuleConfig {
	return m.config
}
func (m *moduleAdapter) Mounts() []module.Mount {
	return m.mounts
}
func (m *moduleAdapter) Owner() module.Module {
	return m.owner
}
func (m *moduleAdapter) IsProj() bool {
	return m.projectMod == true
}

// ApplyProjectConfigDefaults applies default/missing module configuration for
// the main project.
func ApplyProjectConfigDefaults(mod module.Module) {
	projectMod := mod.(*moduleAdapter)

	type dirKeyComponent struct {
		key          string
		component    string
		multilingual bool
	}

	dirKeys := []dirKeyComponent{
		{"contentDir", module.ComponentFolderContent, true},
		{"dataDir", module.ComponentFolderData, false},
		{"layoutDir", module.ComponentFolderLayouts, false},
		{"i18nDir", module.ComponentFolderI18n, false},
		{"archetypeDir", module.ComponentFolderArchetypes, false},
		{"assetDir", module.ComponentFolderAssets, false},
		{"", module.ComponentFolderStatic, false},
	}

	var mounts []module.Mount
	for _, d := range dirKeys {
		if d.multilingual {
			// based on language content configuration
			// multiple language has multiple source folders
			if d.component == module.ComponentFolderContent {
				mounts = append(mounts, module.Mount{
					Lang:   "en",
					Source: "mycontent",
					Target: d.component,
				})
			}
		} else {
			mounts = append(mounts,
				module.Mount{
					Source: d.component,
					Target: d.component,
				})
		}
	}

	projectMod.mounts = mounts
}
