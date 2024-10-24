package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/spf13/afero"
)

type Config struct {
	ConfigSourceFs afero.Fs

	Provider config.Provider

	Root
	Caches
	Security
	Menu
	Module
	Service
	*Language

	Imaging
	MediaType
	OutputFormats

	MinifyC

	Taxonomy
}

func (c *Config) Fs() afero.Fs {
	return c.ConfigSourceFs
}

func (c *Config) Theme() string {
	return c.Root.DefaultTheme()
}

func (c *Config) GetImports(moduleDir string) ([]string, error) {
	var (
		configFilename string
		hasConfigFile  bool
		moduleConfig   valueobject.ModuleConfig
		cfg            config.Provider
		err            error
	)

	moduleConfig = valueobject.EmptyModuleConfig

	configFilename, hasConfigFile = valueobject.CheckConfigFilename(moduleDir, c.ConfigSourceFs)
	if hasConfigFile {
		cfg, err = valueobject.FromFile(c.ConfigSourceFs, configFilename)
		if err != nil {
			return nil, err
		}

		moduleConfig, err = valueobject.DecodeModuleConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	mod := &Module{ModuleConfig: moduleConfig}
	return mod.ImportPaths(), nil
}
