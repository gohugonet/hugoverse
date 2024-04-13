package config

import (
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/spf13/afero"
)

// Format 文件格式类型
type Format string

// TOML 支持的格式，为简单示例，只支持TOML格式
const (
	TOML Format = "toml"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

type Config interface {
	Theme() string
	WorkingDir() string
	PublishDir() string
}

// Provider provides the configuration settings for Hugo.
type Provider interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetParams(key string) maps.Params
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	Get(key string) any
	Set(key string, value any)
	Keys() []string
	Merge(key string, value any)
	SetDefaults(params maps.Params)
	SetDefaultMergeStrategy()
	WalkParams(walkFn func(params ...maps.KeyParams) bool)
	IsSet(key string) bool
}

type SourceDescriptor interface {
	Fs() afero.Fs

	// Filename Path to the config file to use, e.g. /my/project/config.toml
	Filename() string
}

type Compiler interface {
	CompileConfig(logger loggers.Logger) error
}
