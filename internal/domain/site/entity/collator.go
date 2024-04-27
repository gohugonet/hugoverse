package entity

import (
	"golang.org/x/text/collate"
	"sync"
)

type Collator struct {
	sync.Mutex
	c *collate.Collator
}

// CompareStrings compares a and b.
// It returns -1 if a < b, 1 if a > b and 0 if a == b.
// Note that the Collator is not thread safe, so you may want
// to acquire a lock on it before calling this method.
func (c *Collator) CompareStrings(a, b string) int {
	return c.c.CompareString(a, b)
}
