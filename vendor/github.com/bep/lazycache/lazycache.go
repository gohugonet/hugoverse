package lazycache

import (
	"sync"

	"github.com/hashicorp/golang-lru/v2/simplelru"
)

// New creates a new Cache.
func New[K comparable, V any](options Options[K, V]) *Cache[K, V] {
	var onEvict simplelru.EvictCallback[K, *valueWrapper[V]] = nil
	if options.OnEvict != nil {
		onEvict = func(key K, value *valueWrapper[V]) {
			value.wait()
			if value.found {
				options.OnEvict(key, value.value)
			}
		}
	}

	lru, err := simplelru.NewLRU[K, *valueWrapper[V]](int(options.MaxEntries), onEvict)
	if err != nil {
		panic(err)
	}
	c := &Cache[K, V]{
		lru: lru,
	}
	return c
}

// Options holds the cache options.
type Options[K comparable, V any] struct {
	// MaxEntries is the maximum number of entries that the cache should hold.
	// Note that this can also be adjusted after the cache is created with Resize.
	MaxEntries int

	// OnEvict is an optional callback that is called when an entry is evicted.
	OnEvict func(key K, value V)
}

// Cache is a thread-safe resizable LRU cache.
type Cache[K comparable, V any] struct {
	lru *simplelru.LRU[K, *valueWrapper[V]]
	mu  sync.RWMutex

	zerov V
}

// Delete deletes the item with given key from the cache, returning if the
// key was contained.
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Remove(key)
}

// DeleteFunc deletes all entries for which the given function returns true.
func (c *Cache[K, V]) DeleteFunc(matches func(key K, item V) bool) int {
	c.mu.RLock()
	keys := c.lru.Keys()

	var keysToDelete []K
	for _, key := range keys {
		w, _ := c.lru.Peek(key)
		if !w.wait().found {
			continue
		}
		if matches(key, w.value) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	var deleteCount int
	for _, key := range keysToDelete {
		if c.lru.Remove(key) {
			deleteCount++
		}
	}

	return deleteCount
}

// Get returns the value associated with key.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	w := c.get(key)
	c.mu.Unlock()
	if w == nil {
		return c.zerov, false
	}
	w.wait()
	return w.value, w.found
}

// GetOrCreate returns the value associated with key, or creates it if it doesn't.
// It also returns a bool indicating if the value was found in the cache.
// Note that create, the cache prime function, is called once and then not called again for a given key
// unless the cache entry is evicted; it does not block other goroutines from calling GetOrCreate,
// it is not called with the cache lock held.
// Note that any error returned by create will be returned by GetOrCreate and repeated calls with the same key will
// receive the same error.
func (c *Cache[K, V]) GetOrCreate(key K, create func(key K) (V, error)) (V, bool, error) {
	c.mu.Lock()
	w := c.get(key)
	if w != nil {
		c.mu.Unlock()
		w.wait()
		// If w.ready is nil, we will repeat any error from the create function to concurrent callers.
		return w.value, true, w.err
	}

	w = &valueWrapper[V]{
		ready: make(chan struct{}),
	}

	// Concurrent access to the same key will see w, but needs to wait for w.ready
	// to get the value.
	c.lru.Add(key, w)
	c.mu.Unlock()

	isPanic := true

	v, err := func() (v V, err error) {
		defer func() {
			w.err = err
			w.value = v
			w.found = err == nil && !isPanic
			close(w.ready)

			if err != nil || isPanic {
				v = c.zerov
				c.Delete(key)
			}
		}()

		// Create the  value with the lock released.
		v, err = create(key)
		isPanic = false

		return
	}()

	return v, false, err
}

// Resize changes the cache size and returns the number of entries evicted.
func (c *Cache[K, V]) Resize(size int) (evicted int) {
	c.mu.Lock()
	evicted = c.lru.Resize(size)
	c.mu.Unlock()
	return evicted
}

// Set associates value with key.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	c.lru.Add(key, &valueWrapper[V]{value: value, found: true})
	c.mu.Unlock()
}

func (c *Cache[K, V]) get(key K) *valueWrapper[V] {
	w, ok := c.lru.Get(key)
	if !ok {
		return nil
	}
	return w
}

// contains returns true if the given key is in the cache.
// note that this wil also return true if the key is in the cache but the value is not yet ready.
func (c *Cache[K, V]) contains(key K) bool {
	c.mu.RLock()
	b := c.lru.Contains(key)
	c.mu.RUnlock()
	return b
}

// keys returns a slice of the keys in the cache, oldest first.
// note that this wil also include keys that are not yet ready.
func (c *Cache[K, V]) keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Keys()
}

// Len returns the number of items in the cache.
// note that this wil also include values that are not yet ready.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len()
}

// valueWrapper holds a cache value that is not available unless the done channel is nil or closed.
// This construct makes more sense if you look at the code in GetOrCreate.
type valueWrapper[V any] struct {
	value V
	found bool
	err   error
	ready chan struct{}
}

func (w *valueWrapper[V]) wait() *valueWrapper[V] {
	if w.ready != nil {
		<-w.ready
	}
	return w
}
