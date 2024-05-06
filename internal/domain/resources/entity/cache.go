package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"sync"
)

type Cache struct {
	filecache.Caches

	sync.RWMutex

	CacheImage                  *dynacache.Partition[string, *ResourceAdapter]
	CacheResource               *dynacache.Partition[string, resources.Resource]
	CacheResources              *dynacache.Partition[string, []resources.Resource]
	CacheResourceTransformation *dynacache.Partition[string, *ResourceAdapterInner]
}

func (c *Cache) GetOrCreateResource(key string, f func() (resources.Resource, error)) (resources.Resource, error) {
	return c.CacheResource.GetOrCreate(key, func(key string) (resources.Resource, error) {
		return f()
	})
}
