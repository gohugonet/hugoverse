package types

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"html/template"
	"reflect"
)

// ToStringE converts v to a string.
func ToStringE(v any) (string, error) {
	if s, ok := TypeToString(v); ok {
		return s, nil
	}

	switch s := v.(type) {
	case json.RawMessage:
		return string(s), nil
	default:
		return cast.ToStringE(v)
	}
}

// ToString converts v to a string.
func ToString(v any) string {
	s, _ := ToStringE(v)
	return s
}

// TypeToString converts v to a string if it's a valid string type.
// Note that this will not try to convert numeric values etc.,
// use ToString for that.
func TypeToString(v any) (string, bool) {
	switch s := v.(type) {
	case string:
		return s, true
	case template.HTML:
		return string(s), true
	case template.CSS:
		return string(s), true
	case template.HTMLAttr:
		return string(s), true
	case template.JS:
		return string(s), true
	case template.JSStr:
		return string(s), true
	case template.URL:
		return string(s), true
	case template.Srcset:
		return string(s), true
	}

	return "", false
}

// ToStringSlicePreserveString is the same as ToStringSlicePreserveStringE,
// but it never fails.
func ToStringSlicePreserveString(v any) []string {
	vv, _ := ToStringSlicePreserveStringE(v)
	return vv
}

// ToStringSlicePreserveStringE converts v to a string slice.
// If v is a string, it will be wrapped in a string slice.
func ToStringSlicePreserveStringE(v any) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	if sds, ok := v.(string); ok {
		return []string{sds}, nil
	}
	result, err := cast.ToStringSliceE(v)
	if err == nil {
		return result, nil
	}

	// Probably []int or similar. Fall back to reflect.
	vv := reflect.ValueOf(v)

	switch vv.Kind() {
	case reflect.Slice, reflect.Array:
		result = make([]string, vv.Len())
		for i := 0; i < vv.Len(); i++ {
			s, err := cast.ToStringE(vv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			result[i] = s
		}
		return result, nil
	default:
		return nil, fmt.Errorf("failed to convert %T to a string slice", v)
	}

}
