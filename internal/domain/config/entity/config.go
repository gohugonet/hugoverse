package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
)

type Config struct {
	valueobject.DefaultConfigProvider
	languages []config.Language
}

func (c *Config) Languages() []config.Language {
	return c.languages
}

func (c *Config) SetLanguages(languages []config.Language) {
	c.languages = languages
}

func (c *Config) SetRoot(m map[string]any) {
	c.Set("", m)
}

func (c *Config) SetDefault() {
	c.SetDefaults(config.Params{
		"contentDir":             "content",
		"resourceDir":            "resources",
		"publishDir":             "public",
		"publishDirOrig":         "public",
		"themesDir":              "themes",
		"assetDir":               "assets",
		"layoutDir":              "layouts",
		"i18nDir":                "i18n",
		"dataDir":                "data",
		"archetypeDir":           "archetypes",
		"configDir":              "config",
		"staticDir":              "static",
		"timeout":                "30s",
		"defaultContentLanguage": "en",
	})
}
