package lazy

import (
	"sync"
	"sync/atomic"
)

// onceMore is similar to sync.Once.
//
// Additional features are:
// * it can be reset, so the action can be repeated if needed
// * it has methods to check if it's done or in progress
//
type onceMore struct {
	mu   sync.Mutex
	lock uint32
	done uint32
}

func (t *onceMore) Done() bool {
	return atomic.LoadUint32(&t.done) == 1
}

func (t *onceMore) Do(f func()) {
	if atomic.LoadUint32(&t.done) == 1 {
		return
	}

	// f may call this Do and we would get a deadlock.
	locked := atomic.CompareAndSwapUint32(&t.lock, 0, 1)
	if !locked {
		return
	}
	defer atomic.StoreUint32(&t.lock, 0)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Double check
	if t.done == 1 {
		return
	}
	defer atomic.StoreUint32(&t.done, 1)
	f()
}

func (t *onceMore) InProgress() bool {
	return atomic.LoadUint32(&t.lock) == 1
}

func (t *onceMore) ResetWithLock() *sync.Mutex {
	t.mu.Lock()
	defer atomic.StoreUint32(&t.done, 0)
	return &t.mu
}
