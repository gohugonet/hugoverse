package entity

import "sync"

var spc = newPageCache()

type pageCache struct {
	sync.RWMutex
	m map[string][]pageCacheEntry
}

func newPageCache() *pageCache {
	return &pageCache{m: make(map[string][]pageCacheEntry)}
}
func (c *pageCache) clear() {
	c.Lock()
	defer c.Unlock()
	c.m = make(map[string][]pageCacheEntry)
}

func (c *pageCache) get(key string, apply func(p Pages), pageLists ...Pages) (Pages, bool) {
	return c.getP(key, func(p *Pages) {
		if apply != nil {
			apply(*p)
		}
	}, pageLists...)
}

func (c *pageCache) getP(key string, apply func(p *Pages), pageLists ...Pages) (Pages, bool) {
	c.RLock()
	if cached, ok := c.m[key]; ok {
		for _, entry := range cached {
			if entry.matches(pageLists) {
				c.RUnlock()
				return entry.out, true
			}
		}
	}
	c.RUnlock()

	c.Lock()
	defer c.Unlock()

	// double-check
	if cached, ok := c.m[key]; ok {
		for _, entry := range cached {
			if entry.matches(pageLists) {
				return entry.out, true
			}
		}
	}

	p := pageLists[0]
	pagesCopy := append(Pages(nil), p...)

	if apply != nil {
		apply(&pagesCopy)
	}

	entry := pageCacheEntry{in: pageLists, out: pagesCopy}
	if v, ok := c.m[key]; ok {
		c.m[key] = append(v, entry)
	} else {
		c.m[key] = []pageCacheEntry{entry}
	}

	return pagesCopy, false
}

type pageCacheEntry struct {
	in  []Pages
	out Pages
}

func (entry pageCacheEntry) matches(pageLists []Pages) bool {
	if len(entry.in) != len(pageLists) {
		return false
	}
	for i, p := range pageLists {
		if !pagesEqual(p, entry.in[i]) {
			return false
		}
	}

	return true
}

// pagesEqual returns whether p1 and p2 are equal.
func pagesEqual(p1, p2 Pages) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	if p1 == nil || p2 == nil {
		return false
	}

	if p1.Len() != p2.Len() {
		return false
	}

	if p1.Len() == 0 {
		return true
	}

	for i := 0; i < len(p1); i++ {
		if p1[i].PageIdentity() != p2[i].PageIdentity() {
			return false
		}
	}
	return true
}
