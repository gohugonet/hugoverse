package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/config"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"github.com/mdfriday/hugoverse/pkg/media"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"reflect"
	"sort"
	"strings"
)

type MediaTypeConfig struct {
	// The file suffixes used for this media type.
	Suffixes []string
	// Delimiter used before suffix.
	Delimiter string

	Types media.Types
}

func DecodeMediaTypesConfig(p config.Provider) (MediaTypeConfig, error) {
	in := p.GetStringMap("mediatypes")

	buildConfig := func(v any) (media.Types, error) {
		var m map[string]any
		var err error

		if v != nil {
			m, err = maps.ToStringMapE(v)
			if err != nil {
				return nil, err
			}
		}

		if m == nil {
			m = map[string]any{}
		}
		m = maps.CleanConfigStringMap(m)
		// Merge with defaults.
		maps.MergeShallow(m, defaultMediaTypesConfig)

		var types media.Types

		for k, v := range m {
			mediaType, err := media.FromString(k)
			if err != nil {
				return nil, err
			}
			if err := mapstructure.WeakDecode(v, &mediaType); err != nil {
				return nil, err
			}
			mm := maps.ToStringMap(v)
			suffixes, found := maps.LookupEqualFold(mm, "suffixes")
			if found {
				mediaType.SuffixesCSV = strings.TrimSpace(strings.ToLower(strings.Join(cast.ToStringSlice(suffixes), ",")))
			}
			if mediaType.SuffixesCSV != "" && mediaType.Delimiter == "" {
				mediaType.Delimiter = media.DefaultDelimiter
			}
			media.InitMediaType(&mediaType)
			types = append(types, mediaType)
		}

		sort.Sort(types)

		return types, nil
	}

	// Build the config
	c, err := buildConfig(in)
	if err != nil {
		return MediaTypeConfig{}, err
	}

	setupBuildInTypes(c)

	return MediaTypeConfig{
		Suffixes:  nil,
		Delimiter: "",
		Types:     c,
	}, nil
}

func setupBuildInTypes(defaultTypes media.Types) {
	// Initialize the Builtin types with values from DefaultTypes.
	v := reflect.ValueOf(&media.Builtin).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fieldName := v.Type().Field(i).Name
		builtinType := f.Interface().(media.Type)
		if builtinType.Type == "" {
			panic(fmt.Errorf("builtin type %q is empty", fieldName))
		}
		defaultType, found := defaultTypes.GetByType(builtinType.Type)
		if !found {
			panic(fmt.Errorf("missing default type for field builtin type: %q", fieldName))
		}
		f.Set(reflect.ValueOf(defaultType))
	}
}
