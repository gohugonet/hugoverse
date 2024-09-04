package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"strings"
)

type LayoutCacheKey struct {
	Names []string
}

func (k LayoutCacheKey) String() string {
	return strings.Join(k.Names, "/")
}

func (k LayoutCacheKey) IsEmpty() bool {
	return len(k.Names) == 0
}

type LayoutCacheEntry struct {
	Found bool
	Templ template.Preparer
	Err   error
}

func NewLayoutCacheEntry(found bool, templ template.Preparer, err error) LayoutCacheEntry {
	return LayoutCacheEntry{Found: found, Templ: templ, Err: err}
}
