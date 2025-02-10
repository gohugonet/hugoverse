package entity

import (
	"github.com/mdfriday/hugoverse/pkg/paths"
	"strings"
)

func (c *Config) ThemesDir() string {
	// TODO
	return c.Root.DefaultTheme()
}

func (c *Config) WorkingDir() string {
	return c.Root.RootConfig.BaseDirs.WorkingDir
}

func (c *Config) PublishDir() string {
	return c.Root.RootConfig.BaseDirs.PublishDir
}

func (c *Config) ResourceDir() string { return c.Root.RootConfig.ResourceDir }

func (c *Config) AbsoluteResourcesDir() string {
	absResourcesDir := paths.AbsPathify(c.WorkingDir(), c.Root.RootConfig.ResourceDir)
	if !strings.HasSuffix(absResourcesDir, paths.FilePathSeparator) {
		absResourcesDir += paths.FilePathSeparator
	}
	if absResourcesDir == "//" {
		absResourcesDir = paths.FilePathSeparator
	}

	return absResourcesDir
}
