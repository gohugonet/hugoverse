package stale

type Value[V any] struct {
	// The value.
	Value V

	// IsStaleFunc reports whether the value is stale.
	IsStaleFunc func() bool
}

func (s *Value[V]) IsStale() bool {
	return s.IsStaleFunc()
}
