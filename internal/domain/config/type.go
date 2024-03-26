package config

// Format 文件格式类型
type Format string

// TOML 支持的格式，为简单示例，只支持TOML格式
const (
	TOML Format = "toml"
)

type Provider interface {
	Get(key string) any
	Set(key string, value any)
	SetDefaults(params Params)
	GetString(key string) string
	IsSet(key string) bool
}

type Language interface {
	Language() string
	Provider
}

type LanguageProvider interface {
	Languages() []Language
	Provider
}

type Params map[string]any
