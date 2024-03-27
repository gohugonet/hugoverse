package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/entity"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"path"
)

func New() *entity.Config {
	return &entity.Config{
		DefaultConfigProvider: valueobject.DefaultConfigProvider{
			Root: make(config.Params),
		}}
}

func NewConfigFromPath(projPath string) (config.LanguageProvider, error) {
	c := &entity.ConfigLoader{
		Path: path.Join(projPath, "config.toml"),
	}

	m, err := c.LoadConfigFromDisk()
	if err != nil {
		return nil, err
	}

	provider := New()
	provider.SetRoot(m)
	provider.Set("path", projPath)
	provider.Set("workingDir", projPath)
	provider.SetDefault()

	provider.SetLanguages([]config.Language{NewDefaultLanguage(provider)})

	return provider, nil
}

// NewDefaultLanguage creates the default language for a config.Provider.
// If not otherwise specified the default is "en".
func NewDefaultLanguage(cfg config.Provider) *valueobject.Language {
	defaultLang := cfg.GetString("defaultContentLanguage")

	if defaultLang == "" {
		defaultLang = "en"
	}

	return NewLanguage(defaultLang, cfg)
}

// NewLanguage creates a new language.
func NewLanguage(lang string, cfg config.Provider) *valueobject.Language {
	localCfg := New()
	compositeConfig := NewCompositeConfig(cfg, localCfg)

	l := &valueobject.Language{
		Lang:       lang,
		ContentDir: cfg.GetString("contentDir"),
		Cfg:        cfg,
		LocalCfg:   localCfg,
		Provider:   compositeConfig,
	}

	return l
}

// NewCompositeConfig creates a new composite Provider with a read-only base
// and a writeable layer.
func NewCompositeConfig(base, layer config.Provider) config.Provider {
	return &valueobject.CompositeConfig{
		Base:  base,
		Layer: layer,
	}
}
