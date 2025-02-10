package valueobject

import (
	"github.com/mdfriday/hugoverse/internal/domain/config"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
)

type ModuleConfig struct {
	// File system mounts.
	Mounts []Mount

	// Module imports.
	Imports []Import
}

// DecodeModuleConfig creates a modules Config from a given Hugo configuration.
func DecodeModuleConfig(cfg config.Provider) (ModuleConfig, error) {
	return decodeConfig(cfg)
}

func decodeConfig(cfg config.Provider) (ModuleConfig, error) {
	c := EmptyModuleConfig

	moduleSet := cfg.IsSet("module")
	if moduleSet {
		m := cfg.GetStringMap("module")
		if err := mapstructure.WeakDecode(m, &c); err != nil {
			return c, err
		}

		for i, mnt := range c.Mounts {
			mnt.Source = filepath.Clean(mnt.Source)
			mnt.Target = filepath.Clean(mnt.Target)
			c.Mounts[i] = mnt
		}
	}

	return c, nil
}

var EmptyModuleConfig = ModuleConfig{
	Mounts:  make([]Mount, 0),
	Imports: make([]Import, 0),
}
