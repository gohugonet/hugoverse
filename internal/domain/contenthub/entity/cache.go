package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/cache/dynacache"
	"github.com/mdfriday/hugoverse/pkg/cache/stale"
)

type Cache struct {
	CachePages1 *dynacache.Partition[string, contenthub.Pages]
	CachePages2 *dynacache.Partition[string, contenthub.Pages]

	// Cache for content sources.
	CacheContentSource *dynacache.Partition[string, *stale.Value[[]byte]]

	CachePageSource  *dynacache.Partition[string, contenthub.PageSource]
	CachePageSources *dynacache.Partition[string, []contenthub.PageSource]

	CacheContentRendered   *dynacache.Partition[string, *stale.Value[valueobject.ContentSummary]]
	CacheContentToRender   *dynacache.Partition[string, *stale.Value[[]byte]]
	CacheContentShortcodes *dynacache.Partition[string, *stale.Value[map[string]valueobject.ShortcodeRenderer]]
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
