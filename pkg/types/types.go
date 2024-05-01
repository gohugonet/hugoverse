package types

import "reflect"

// IsNil reports whether v is nil.
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	}

	return false
}

// Zeroer, as implemented by time.Time, will be used by the truth template
// funcs in Hugo (if, with, not, and, or).
type Zeroer interface {
	IsZero() bool
}

// RLocker represents the read locks in sync.RWMutex.
type RLocker interface {
	RLock()
	RUnlock()
}

// DevMarker is a marker interface for types that should only be used during
// development.
type DevMarker interface {
	DevOnly()
}

// Unwrapper is implemented by types that can unwrap themselves.
type Unwrapper interface {
	// Unwrapv is for internal use only.
	// It got its slightly odd name to prevent collisions with user types.
	Unwrapv() any
}

// Unwrap returns the underlying value of v if it implements Unwrapper, otherwise v is returned.
func Unwrapv(v any) any {
	if u, ok := v.(Unwrapper); ok {
		return u.Unwrapv()
	}
	return v
}

// LowHigh is typically used to represent a slice boundary.
type LowHigh struct {
	Low  int
	High int
}
