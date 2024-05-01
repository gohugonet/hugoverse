package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
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
		m, err := maps.ToStringMapE(v)
		if err != nil {
			return nil, err
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

	return MediaTypeConfig{
		Suffixes:  nil,
		Delimiter: "",
		Types:     c,
	}, nil
}
