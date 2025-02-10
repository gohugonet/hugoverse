package stale

import "github.com/mdfriday/hugoverse/pkg/types"

// Staler controls stale state of a Resource. A stale resource should be discarded.
type Staler interface {
	Marker
	Info
}

// Marker marks a Resource as stale.
type Marker interface {
	MarkStale()
}

// Info tells if a resource is marked as stale.
type Info interface {
	IsStale() bool
}

// IsStaleAny reports whether any of the os is marked as stale.
func IsStaleAny(os ...any) bool {
	for _, o := range os {
		if s, ok := o.(Info); ok && s.IsStale() {
			return true
		}
	}
	return false
}

// MarkStale will mark any of the oses as stale, if possible.
func MarkStale(os ...any) {
	for _, o := range os {
		if types.IsNil(o) {
			continue
		}
		if s, ok := o.(Marker); ok {
			s.MarkStale()
		}
	}
}
