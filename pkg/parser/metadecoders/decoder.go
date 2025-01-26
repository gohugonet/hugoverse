package metadecoders

import (
	"encoding/json"
	"fmt"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
	"strings"
)

// Decoder provides some configuration options for the decoders.
type Decoder struct {
	// Delimiter is the field delimiter used in the CSV decoder. It defaults to ','.
	Delimiter rune

	// Comment, if not 0, is the comment character ued in the CSV decoder. Lines beginning with the
	// Comment character without preceding whitespace are ignored.
	Comment rune
}

// Default is a Decoder in its default configuration.
var Default = Decoder{
	Delimiter: ',',
}

// UnmarshalFileToMap is the same as UnmarshalToMap, but reads the data from
// the given filename.
func (d Decoder) UnmarshalFileToMap(fs afero.Fs, filename string) (map[string]any, error) {
	format := FormatFromString(filename)
	if format == "" {
		return nil, fmt.Errorf("%q is not a valid configuration format", filename)
	}

	data, err := afero.ReadFile(fs, filename)
	if err != nil {
		return nil, err
	}
	return d.UnmarshalToMap(data, format)
}

// UnmarshalToMap will unmarshall data in format f into a new map. This is
// what's needed for Hugo's front matter decoding.
func (d Decoder) UnmarshalToMap(data []byte, f Format) (map[string]any, error) {
	m := make(map[string]any)
	if data == nil {
		return m, nil
	}

	err := d.UnmarshalTo(data, f, &m)

	return m, err
}

// UnmarshalTo unmarshals data in format f into v.
func (d Decoder) UnmarshalTo(data []byte, f Format, v any) error {
	var err error

	switch f {
	case TOML:
		err = toml.Unmarshal(data, v)
	case JSON:
		err = json.Unmarshal(data, v)
	case YAML:
		err = yaml.Unmarshal(data, v)
		if err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %w, %s", err, string(data))
		}

		// To support boolean keys, the YAML package unmarshals maps to
		// map[interface{}]interface{}. Here we recurse through the result
		// and change all maps to map[string]interface{} like we would've
		// gotten from `json`.
		var ptr any
		switch v.(type) {
		case *map[string]any:
			ptr = *v.(*map[string]any)
		case *any:
			ptr = *v.(*any)
		default:
			// Not a map.
		}

		if ptr != nil {
			if mm, changed := stringifyMapKeys(ptr); changed {
				switch v.(type) {
				case *map[string]any:
					*v.(*map[string]any) = mm.(map[string]any)
				case *any:
					*v.(*any) = mm
				}
			}
		}
	default:
		return fmt.Errorf("unmarshal of format %q is not supported", f)
	}

	if err == nil {
		return nil
	}

	return err
}

// Unmarshal will unmarshall data in format f into an interface{}.
// This is what's needed for Hugo's /data handling.
func (d Decoder) Unmarshal(data []byte, f Format) (any, error) {
	if data == nil {
		switch f {
		default:
			return make(map[string]any), nil
		}
	}
	var v any
	err := d.UnmarshalTo(data, f, &v)

	return v, err
}

// OptionsKey is used in cache keys.
func (d Decoder) OptionsKey() string {
	var sb strings.Builder
	sb.WriteRune(d.Delimiter)
	sb.WriteRune(d.Comment)
	return sb.String()
}

// stringifyMapKeys recurses into in and changes all instances of
// map[interface{}]interface{} to map[string]interface{}. This is useful to
// work around the impedance mismatch between JSON and YAML unmarshaling that's
// described here: https://github.com/go-yaml/yaml/issues/139
//
// Inspired by https://github.com/stripe/stripe-mock, MIT licensed
func stringifyMapKeys(in any) (any, bool) {
	switch in := in.(type) {
	case []any:
		for i, v := range in {
			if vv, replaced := stringifyMapKeys(v); replaced {
				in[i] = vv
			}
		}
	case map[string]any:
		for k, v := range in {
			if vv, changed := stringifyMapKeys(v); changed {
				in[k] = vv
			}
		}
	case map[any]any:
		res := make(map[string]any)
		var (
			ok  bool
			err error
		)
		for k, v := range in {
			var ks string

			if ks, ok = k.(string); !ok {
				ks, err = cast.ToStringE(k)
				if err != nil {
					ks = fmt.Sprintf("%v", k)
				}
			}
			if vv, replaced := stringifyMapKeys(v); replaced {
				res[ks] = vv
			} else {
				res[ks] = v
			}
		}
		return res, true
	}

	return nil, false
}
