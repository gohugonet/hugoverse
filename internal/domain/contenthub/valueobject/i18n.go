package valueobject

import (
	"github.com/mdfriday/hugoverse/pkg/hreflect"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

// IntCount wraps the Count method.
type IntCount int

func (c IntCount) Count() int {
	return int(c)
}

func GetPluralCount(v any) any {
	if v == nil {
		// i18n called without any argument, make sure it does not
		// get any plural count.
		return nil
	}

	switch v := v.(type) {
	case map[string]any:
		for k, vv := range v {
			if strings.EqualFold(k, countFieldName) {
				return toPluralCountValue(vv)
			}
		}
	default:
		vv := reflect.Indirect(reflect.ValueOf(v))
		if vv.Kind() == reflect.Interface && !vv.IsNil() {
			vv = vv.Elem()
		}
		tp := vv.Type()

		if tp.Kind() == reflect.Struct {
			f := vv.FieldByName(countFieldName)
			if f.IsValid() {
				return toPluralCountValue(f.Interface())
			}
			m := hreflect.GetMethodByName(vv, countFieldName)
			if m.IsValid() && m.Type().NumIn() == 0 && m.Type().NumOut() == 1 {
				c := m.Call(nil)
				return toPluralCountValue(c[0].Interface())
			}
		}
	}

	return toPluralCountValue(v)
}

const countFieldName = "Count"

// go-i18n expects floats to be represented by string.
func toPluralCountValue(in any) any {
	k := reflect.TypeOf(in).Kind()
	switch {
	case hreflect.IsFloat(k):
		f := cast.ToString(in)
		if !strings.Contains(f, ".") {
			f += ".0"
		}
		return f
	case k == reflect.String:
		if _, err := cast.ToFloat64E(in); err == nil {
			return in
		}
		// A non-numeric value.
		return nil
	default:
		if i, err := cast.ToIntE(in); err == nil {
			return i
		}
		return nil
	}
}
