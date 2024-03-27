package lazy

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Init holds a graph of lazily initialized dependencies.
type Init struct {
	// Used in tests
	initCount uint64

	mu sync.Mutex

	prev     *Init
	children []*Init

	init onceMore
	out  any
	err  error
	f    func() (any, error)
}

// New creates a new empty Init.
func New() *Init {
	return &Init{}
}

// Branch creates a new dependency branch based on an existing and adds
// the given dependency as a child.
func (ini *Init) Branch(initFn func() (any, error)) *Init {
	if ini == nil {
		ini = New()
	}
	return ini.add(true, initFn)
}

func (ini *Init) add(branch bool, initFn func() (any, error)) *Init {
	ini.mu.Lock()
	defer ini.mu.Unlock()

	if branch {
		return &Init{
			f:    initFn,
			prev: ini,
		}
	}

	ini.checkDone()
	ini.children = append(ini.children, &Init{
		f: initFn,
	})

	return ini
}

func (ini *Init) checkDone() {
	if ini.init.Done() {
		panic("init cannot be added to after it has run")
	}
}

// Do initializes the entire dependency graph.
func (ini *Init) Do() (any, error) {
	if ini == nil {
		panic("init is nil")
	}

	ini.init.Do(func() {
		atomic.AddUint64(&ini.initCount, 1)
		prev := ini.prev
		if prev != nil {
			// A branch. Initialize the ancestors.
			if prev.shouldInitialize() {
				_, err := prev.Do()
				if err != nil {
					ini.err = err
					return
				}
			} else if prev.inProgress() {
				// Concurrent initialization. The following init func
				// may depend on earlier state, so wait.
				prev.wait()
			}
		}

		if ini.f != nil {
			ini.out, ini.err = ini.f()
		}

		for _, child := range ini.children {
			if child.shouldInitialize() {
				_, err := child.Do()
				if err != nil {
					ini.err = err
					return
				}
			}
		}
	})

	ini.wait()

	return ini.out, ini.err
}

func (ini *Init) shouldInitialize() bool {
	return !(ini == nil || ini.init.Done() || ini.init.InProgress())
}

func (ini *Init) inProgress() bool {
	return ini != nil && ini.init.InProgress()
}

// TODO(bep) investigate if we can use sync.Cond for this.
func (ini *Init) wait() {
	var counter time.Duration
	for !ini.init.Done() {
		counter += 10
		if counter > 600000000 {
			panic("BUG: timed out in lazy init")
		}
		time.Sleep(counter * time.Microsecond)
	}
}

// Add adds a func as a new child dependency.
func (ini *Init) Add(initFn func() (any, error)) *Init {
	if ini == nil {
		ini = New()
	}
	return ini.add(false, initFn)
}

// BranchdWithTimeout is same as Branch, but with a timeout.
func (ini *Init) BranchWithTimeout(timeout time.Duration, f func(ctx context.Context) (any, error)) *Init {
	return ini.Branch(func() (any, error) {
		return ini.withTimeout(timeout, f)
	})
}

func (ini *Init) withTimeout(timeout time.Duration, f func(ctx context.Context) (any, error)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := make(chan verr, 1)

	go func() {
		v, err := f(ctx)
		select {
		case <-ctx.Done():
			return
		default:
			c <- verr{v: v, err: err}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("timed out initializing value. You may have a circular loop in a shortcode, or your site may have resources that take longer to build than the `timeout` limit in your Hugo config file.")
	case ve := <-c:
		return ve.v, ve.err
	}
}

type verr struct {
	v   any
	err error
}

// Reset resets the current and all its dependencies.
func (ini *Init) Reset() {
	mu := ini.init.ResetWithLock()
	ini.err = nil
	defer mu.Unlock()
	for _, d := range ini.children {
		d.Reset()
	}
}
