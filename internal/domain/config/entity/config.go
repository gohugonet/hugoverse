package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
)

type Config struct {
	Provider config.Provider

	Root
	Module
	Language
}

func (c *Config) Theme() string {
	return c.Root.DefaultTheme()
}

func (c *Config) WorkingDir() string {
	return c.Root.WorkingDir
}

func (c *Config) PublishDir() string {
	return c.Root.PublishDir
}
