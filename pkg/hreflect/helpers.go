package hreflect

import (
	"github.com/gohugonet/hugoverse/pkg/htime"
	"github.com/gohugonet/hugoverse/pkg/types"
	"reflect"
	"sync"
	"time"
)

// GetMethodByName is the same as reflect.Value.MethodByName, but it caches the
// type lookup.
func GetMethodByName(v reflect.Value, name string) reflect.Value {
	index := GetMethodIndexByName(v.Type(), name)

	if index == -1 {
		return reflect.Value{}
	}

	return v.Method(index)
}

var methodCache = &methods{cache: make(map[methodKey]int)}

type methods struct {
	sync.RWMutex
	cache map[methodKey]int
}

// GetMethodIndexByName returns the index of the method with the given name, or
// -1 if no such method exists.
func GetMethodIndexByName(tp reflect.Type, name string) int {
	k := methodKey{tp, name}
	methodCache.RLock()
	index, found := methodCache.cache[k]
	methodCache.RUnlock()
	if found {
		return index
	}

	methodCache.Lock()
	defer methodCache.Unlock()

	m, ok := tp.MethodByName(name)
	index = m.Index
	if !ok {
		index = -1
	}
	methodCache.cache[k] = index

	if !ok {
		return -1
	}

	return m.Index
}

type methodKey struct {
	typ  reflect.Type
	name string
}

var zeroType = reflect.TypeOf((*types.Zeroer)(nil)).Elem()

// IsTruthfulValue returns whether the given value has a meaningful truth value.
// This is based on template.IsTrue in Go's stdlib, but also considers
// IsZero and any interface value will be unwrapped before it's considered
// for truthfulness.
//
// Based on:
// https://github.com/golang/go/blob/178a2c42254166cffed1b25fb1d3c7a5727cada6/src/text/template/exec.go#L306
func IsTruthfulValue(val reflect.Value) (truth bool) {
	val = indirectInterface(val)

	if !val.IsValid() {
		// Something like var x interface{}, never set. It's a form of nil.
		return
	}

	if val.Type().Implements(zeroType) {
		return !val.Interface().(types.Zeroer).IsZero()
	}

	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		truth = val.Len() > 0
	case reflect.Bool:
		truth = val.Bool()
	case reflect.Complex64, reflect.Complex128:
		truth = val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		truth = !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		truth = val.Int() != 0
	case reflect.Float32, reflect.Float64:
		truth = val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		truth = val.Uint() != 0
	case reflect.Struct:
		truth = true // Struct values are always true.
	default:
		return
	}

	return
}

// Based on: https://github.com/golang/go/blob/178a2c42254166cffed1b25fb1d3c7a5727cada6/src/text/template/exec.go#L931
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

var (
	timeType           = reflect.TypeOf((*time.Time)(nil)).Elem()
	asTimeProviderType = reflect.TypeOf((*htime.AsTimeProvider)(nil)).Elem()
)

// IsTime returns whether tp is a time.Time type or if it can be converted into one
// in ToTime.
func IsTime(tp reflect.Type) bool {
	if tp == timeType {
		return true
	}

	if tp.Implements(asTimeProviderType) {
		return true
	}
	return false
}

// AsTime returns v as a time.Time if possible.
// The given location is only used if the value implements AsTimeProvider (e.g. go-toml local).
// A zero Time and false is returned if this isn't possible.
// Note that this function does not accept string dates.
func AsTime(v reflect.Value, loc *time.Location) (time.Time, bool) {
	if v.Kind() == reflect.Interface {
		return AsTime(v.Elem(), loc)
	}

	if v.Type() == timeType {
		return v.Interface().(time.Time), true
	}

	if v.Type().Implements(asTimeProviderType) {
		return v.Interface().(htime.AsTimeProvider).AsTime(loc), true
	}

	return time.Time{}, false
}

// IsFloat returns whether the given kind is a float.
func IsFloat(kind reflect.Kind) bool {
	switch kind {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsInt returns whether the given kind is an int.
func IsInt(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}
