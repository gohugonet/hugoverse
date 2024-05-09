package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"sort"
	"strings"
)

type OutputFormatsConfig struct {
	Configs map[string]OutputFormatConfig
	output.Formats
}

// OutputFormatConfig configures a single output format.
type OutputFormatConfig struct {
	// The MediaType string. This must be a configured media type.
	MediaType string
	output.Format
}

var defaultOutputFormat = output.Format{
	BaseName: "index",
	Rel:      "alternate",
}

func DecodeOutputFormatConfig(mediaTypes media.Types, p config.Provider) (OutputFormatsConfig, error) {
	in := p.GetStringMap("outputformats")

	buildConfig := func(in any) (output.Formats, map[string]OutputFormatConfig, error) {
		f := make(output.Formats, len(output.DefaultFormats))
		copy(f, output.DefaultFormats)

		if in != nil {
			m, err := maps.ToStringMapE(in)
			if err != nil {
				return nil, nil, fmt.Errorf("failed convert config to map: %s", err)
			}
			m = maps.CleanConfigStringMap(m)

			for k, v := range m {
				found := false
				for i, vv := range f {
					// Both are lower case.
					if k == vv.Name {
						// Merge it with the existing
						if err := decodeOutputFormat(mediaTypes, v, &f[i]); err != nil {
							return f, nil, err
						}
						found = true
					}
				}
				if found {
					continue
				}

				newOutFormat := defaultOutputFormat
				if err := decodeOutputFormat(mediaTypes, v, &newOutFormat); err != nil {
					return f, nil, err
				}
				newOutFormat.Name = k

				f = append(f, newOutFormat)

			}
		}

		// Also format is a map for documentation purposes.
		docm := make(map[string]OutputFormatConfig, len(f))
		for _, ff := range f {
			docm[ff.Name] = OutputFormatConfig{
				MediaType: ff.MediaType.Type,
				Format:    ff,
			}
		}

		sort.Sort(f)
		return f, docm, nil
	}

	f, configs, err := buildConfig(in)
	if err != nil {
		return OutputFormatsConfig{}, err
	}

	return OutputFormatsConfig{
		Configs: configs,
		Formats: f,
	}, nil
}

func decodeOutputFormat(mediaTypes media.Types, input any, output *output.Format) error {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook: func(a reflect.Type, b reflect.Type, c any) (any, error) {
			if a.Kind() == reflect.Map {
				dataVal := reflect.Indirect(reflect.ValueOf(c))
				for _, key := range dataVal.MapKeys() {
					keyStr, ok := key.Interface().(string)
					if !ok {
						// Not a string key
						continue
					}
					if strings.EqualFold(keyStr, "mediaType") {
						// If mediaType is a string, look it up and replace it
						// in the map.
						vv := dataVal.MapIndex(key)
						vvi := vv.Interface()

						switch vviv := vvi.(type) {
						case media.Type:
						// OK
						case string:
							mediaType, found := mediaTypes.GetByType(vviv)
							if !found {
								return c, fmt.Errorf("media type %q not found", vviv)
							}
							dataVal.SetMapIndex(key, reflect.ValueOf(mediaType))
						default:
							return nil, fmt.Errorf("invalid output format configuration; wrong type for media type, expected string (e.g. text/html), got %T", vvi)
						}
					}
				}
			}
			return c, nil
		},
	}

	decoder, err := mapstructure.NewDecoder(c)
	if err != nil {
		return err
	}

	if err = decoder.Decode(input); err != nil {
		return fmt.Errorf("failed to decode output format configuration: %w", err)
	}

	return nil
}
