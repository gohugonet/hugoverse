package valueobject

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
)

var DefaultLanguage = LanguageConfig{
	LanguageName:      "English",
	LanguageCode:      "en",
	Title:             "",
	LanguageDirection: "",
	Weight:            0,
	Disabled:          false,
}

// LanguageConfig holds the configuration for a single language.
// This is what is read from the config file.
type LanguageConfig struct {
	// The language name, e.g. "English".
	LanguageName string

	// The language code, e.g. "en-US".
	LanguageCode string

	// The language title. When set, this will
	// override site.Title for this language.
	Title string

	// The language direction, e.g. "ltr" or "rtl".
	LanguageDirection string

	// The language weight. When set to a non-zero value, this will
	// be the main sort criteria for the language.
	Weight int

	// Set to true to disable this language.
	Disabled bool
}

func (l LanguageConfig) Name() string {
	return l.LanguageName
}

func (l LanguageConfig) Code() string {
	return l.LanguageCode
}

func DecodeLanguageConfig(p config.Provider) (map[string]LanguageConfig, error) {
	var err error
	m := p.GetStringMap("languages")
	if len(m) == 1 {
		var first maps.Params
		var ok bool
		for _, v := range m {
			first, ok = v.(maps.Params)
			if ok {
				break
			}
		}
		if first != nil {
			if _, found := first["languagecode"]; !found {
				code := p.GetString("languagecode")
				if code == "" {
					code = "en-US"
				}
				first["languagecode"] = code
			}
		}
	}

	languages, err := decodeLanguageConfig(m)
	if err != nil {
		return nil, err
	}

	return languages, nil
}

func decodeLanguageConfig(m map[string]any) (map[string]LanguageConfig, error) {
	m = maps.CleanConfigStringMap(m)
	var langs map[string]LanguageConfig

	if err := mapstructure.WeakDecode(m, &langs); err != nil {
		return nil, err
	}
	if len(langs) == 0 {
		return nil, errors.New("no languages configured")
	}
	return langs, nil
}
