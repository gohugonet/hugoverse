package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
)

type Language struct {
	Default     string
	Configs     map[string]valueobject.LanguageConfig
	RootConfigs map[string]valueobject.RootConfig
}

func (l Language) GetDefaultDirs(names []string) ([]module.Component, error) {
	var components []module.Component

	root, ok := l.RootConfigs[l.Default]
	if !ok {
		return nil, l.defaultLangError()
	}
	for _, name := range names {
		components = append(components, &valueobject.Component{
			ComName: name,
			ComDir:  root.CommonDirs.GetDirectoryByName(name),
			ComLang: l.Default,
		})
	}

	return components, nil
}

func (l Language) GetOtherLanguagesContentDirs(name string) ([]module.Component, error) {
	var components []module.Component
	for lang, config := range l.RootConfigs {
		if lang == l.Default {
			continue
		}

		components = append(components, &valueobject.Component{
			ComName: name,
			ComDir:  config.CommonDirs.GetDirectoryByName(name),
			ComLang: lang,
		})
	}

	return components, nil
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
