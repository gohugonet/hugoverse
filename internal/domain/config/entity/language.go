package entity

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"golang.org/x/exp/maps"
)

type Language struct {
	Default     string
	Configs     map[string]valueobject.LanguageConfig
	RootConfigs map[string]valueobject.RootConfig

	indices []string
}

func (l *Language) Languages() []valueobject.LanguageConfig {
	return maps.Values(l.Configs)
}

func (l *Language) DefaultLanguage() string {
	return l.Default
}

func (l *Language) IsLanguageValid(lang string) bool {
	_, found := l.Configs[lang]
	return found
}

func (l *Language) OtherLanguageKeys() []string {
	var keys []string
	for k := range l.Configs {
		if k != l.Default {
			keys = append(keys, k)
		}
	}
	return keys
}

func (l *Language) GetRelDir(name string, langKey string) (dir string, err error) {
	root, ok := l.RootConfigs[langKey]
	if !ok {
		return "", fmt.Errorf("language %q not found", langKey)
	}

	return root.CommonDirs.GetDirectoryByName(name), nil
}

func (l *Language) Validate() error {
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

func (l *Language) defaultLangError() error {
	return fmt.Errorf("config value %q for defaultContentLanguage does not match any language definition", l.Default)
}

func (l *Language) SetIndices() {
	var languages []string
	// Ensure default language is first
	if _, exists := l.Configs[l.Default]; exists {
		languages = append(languages, l.Default)
	}
	// Add remaining languages
	for lang := range l.Configs {
		if lang != l.Default {
			languages = append(languages, lang)
		}
	}

	l.indices = languages
}

func (l *Language) LanguageKeys() []string {
	return l.indices
}

func (l *Language) LanguageIndexes() []int {
	var indexes []int
	for i, _ := range l.indices {
		indexes = append(indexes, i)
	}
	return indexes
}

func (l *Language) GetLanguageIndex(lang string) (int, error) {
	for i, v := range l.indices {
		if v == lang {
			return i, nil
		}
	}
	return -1, errors.New("language not found in indices")
}

func (l *Language) GetLanguageByIndex(idx int) string {
	return l.indices[idx]
}

func (l *Language) GetLanguageName(lang string) string {
	for c, v := range l.Configs {
		if c == lang {
			return v.Name()
		}
	}

	return ""
}

func (l *Language) GetLanguageFolder(lang string) string {
	for c, v := range l.RootConfigs {
		if c == lang {
			return v.ContentDir
		}
	}

	return ""
}
