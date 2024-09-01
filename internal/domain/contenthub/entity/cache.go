package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
)

type Cache struct {
	// Cache for content sources.
	CacheContentSource *dynacache.Partition[string, *stale.Value[[]byte]]

	CachePageSource  *dynacache.Partition[string, contenthub.PageSource]
	CachePageSources *dynacache.Partition[string, []contenthub.PageSource]

	CacheContentRendered   *dynacache.Partition[string, *stale.Value[valueobject.ContentSummary]]
	ContentTableOfContents *dynacache.Partition[string, *stale.Value[valueobject.ContentToC]]
}

func (c *Cache) GetOrCreateResource(key string, f func() (contenthub.PageSource, error)) (contenthub.PageSource, error) {
	return c.CachePageSource.GetOrCreate(key, func(key string) (contenthub.PageSource, error) {
		return f()
	})
}

func (c *Cache) GetOrCreateResources(key string, f func() ([]contenthub.PageSource, error)) ([]contenthub.PageSource, error) {
	return c.CachePageSources.GetOrCreate(key, func(key string) ([]contenthub.PageSource, error) {
		return f()
	})
}
