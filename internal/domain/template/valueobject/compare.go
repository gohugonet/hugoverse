package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/compare"
	"github.com/gohugonet/hugoverse/pkg/types"
	"reflect"
)

// Compare provides template functions for the "compare" namespace.
type Compare struct {
}

// Eq returns the boolean truth of arg1 == arg2 || arg1 == arg3 || arg1 == arg4.
func (n *Compare) Eq(first any, others ...any) bool {
	n.checkComparisonArgCount(1, others...)
	normalize := func(v any) any {
		if types.IsNil(v) {
			return nil
		}

		vv := reflect.ValueOf(v)
		switch vv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vv.Int()
		case reflect.Float32, reflect.Float64:
			return vv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vv.Uint()
		case reflect.String:
			return vv.String()
		default:
			return v
		}
	}

	normFirst := normalize(first)
	for _, other := range others {
		if e, ok := first.(compare.Eqer); ok {
			if e.Eq(other) {
				return true
			}
			continue
		}

		if e, ok := other.(compare.Eqer); ok {
			if e.Eq(first) {
				return true
			}
			continue
		}

		other = normalize(other)
		if reflect.DeepEqual(normFirst, other) {
			return true
		}
	}

	return false
}

func (n *Compare) checkComparisonArgCount(min int, others ...any) bool {
	if len(others) < min {
		panic("missing arguments for comparison")
	}
	return true
}
