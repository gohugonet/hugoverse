package entity

import "github.com/gohugonet/hugoverse/internal/domain/config/valueobject"

type Config struct {
	Base              *valueobject.BaseConfig
	LanguageConfigMap map[string]*valueobject.BaseConfig
}

func (c *Config) Theme() string {
	return c.Base.DefaultTheme()
}
func (c *Config) WorkingDir() string {
	return c.Base.WorkingDir
}
func (c *Config) PublishDir() string {
	return c.Base.PublishDir
}
