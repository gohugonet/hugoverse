package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"golang.org/x/exp/maps"
)

type Language struct {
	Default     string
	Configs     map[string]valueobject.LanguageConfig
	RootConfigs map[string]valueobject.RootConfig
}

func (l Language) Languages() []valueobject.LanguageConfig {
	return maps.Values(l.Configs)
}

func (l Language) DefaultLanguageKey() string {
	return l.Default
}

func (l Language) OtherLanguageKeys() []string {
	var keys []string
	for k := range l.Configs {
		if k != l.Default {
			keys = append(keys, k)
		}
	}
	return keys
}

func (l Language) GetRelDir(name string, langKey string) (dir string, err error) {
	root, ok := l.RootConfigs[langKey]
	if !ok {
		return "", fmt.Errorf("language %q not found", langKey)
	}

	return root.CommonDirs.GetDirectoryByName(name), nil
}

func (l Language) Validate() error {
	var found bool
	for lang := range l.Configs {
		if lang == l.Default {
			found = true
			break
		}
	}
	if !found {
		return l.defaultLangError()
	}
	return nil
}

func (l Language) defaultLangError() error {
	return fmt.Errorf("config value %q for defaultContentLanguage does not match any language definition", l.Default)
}
