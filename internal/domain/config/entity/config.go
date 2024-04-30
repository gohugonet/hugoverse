package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/spf13/afero"
)

type Config struct {
	SourceFs afero.Fs

	Provider config.Provider

	Root
	Caches
	Security
	Module
	Language

	Imaging
}

func (c *Config) Fs() afero.Fs {
	return c.SourceFs
}

func (c *Config) Theme() string {
	return c.Root.DefaultTheme()
}

func (c *Config) GetImports(moduleDir string) ([]string, error) {
	var (
		configFilename string
		hasConfigFile  bool
		cfg            config.Provider
		err            error
	)

	configFilename, hasConfigFile = valueobject.CheckConfigFilename(moduleDir, c.SourceFs)
	if hasConfigFile {
		cfg, err = valueobject.FromFile(c.SourceFs, configFilename)
		if err != nil {
			return nil, err
		}
	}

	moduleConfig, err := valueobject.DecodeModuleConfig(cfg)
	if err != nil {
		return nil, err
	}

	mod := &Module{ModuleConfig: moduleConfig}
	return mod.ImportPaths(), nil
}
