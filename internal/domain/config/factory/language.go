package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"reflect"
	"strings"
)

// All non-params config keys for language.
var configLanguageKeys map[string]bool

func init() {
	skip := map[string]bool{
		"internal":   true,
		"c":          true,
		"rootconfig": true,
	}
	configLanguageKeys = make(map[string]bool)
	addKeys := func(v reflect.Value) {
		for i := 0; i < v.NumField(); i++ {
			name := strings.ToLower(v.Type().Field(i).Name)
			if skip[name] {
				continue
			}
			configLanguageKeys[name] = true
		}
	}
	addKeys(reflect.ValueOf(valueobject.BaseConfig{}))
	addKeys(reflect.ValueOf(valueobject.RootConfig{}))
	addKeys(reflect.ValueOf(valueobject.CommonDirs{}))
	addKeys(reflect.ValueOf(valueobject.LanguageConfig{}))
}
