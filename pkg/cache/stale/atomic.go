package stale

import "sync/atomic"

type AtomicStaler struct {
	stale uint32
}

func (s *AtomicStaler) MarkStale() {
	atomic.StoreUint32(&s.stale, 1)
}

func (s *AtomicStaler) IsStale() bool {
	return atomic.LoadUint32(&(s.stale)) > 0
}
