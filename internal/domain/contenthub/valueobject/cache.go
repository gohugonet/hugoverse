package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

type Cache struct {
	// The memory cache to use.
	memCache *dynacache.Cache

	// Cache for content sources.
	CacheContentSource *dynacache.Partition[string, *stale.Value[[]byte]]
}

func NewCache() *Cache {
	memCache := dynacache.New(dynacache.Options{Running: true, Log: loggers.NewDefault()})

	return &Cache{
		memCache: memCache,
		CacheContentSource: dynacache.GetOrCreatePartition[string, *stale.Value[[]byte]](
			memCache, "/cont/src",
			dynacache.OptionsPartition{Weight: 70, ClearWhen: dynacache.ClearOnChange},
		),
	}
}
