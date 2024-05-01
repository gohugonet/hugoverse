package valueobject

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"io"
	"sync"
)

type ResourceCache struct {
	sync.RWMutex

	CacheResource               *dynacache.Partition[string, resources.Resource]
	CacheResources              *dynacache.Partition[string, []resources.Resource]
	CacheResourceTransformation *dynacache.Partition[string, *ResourceAdapterInner]

	FileCache *filecache.Cache
}

func (c *ResourceCache) GetOrCreate(key string, f func() (resources.Resource, error)) (resources.Resource, error) {
	return c.CacheResource.GetOrCreate(key, func(key string) (resources.Resource, error) {
		return f()
	})
}

func (c *ResourceCache) GetFromFile(key string) (filecache.ItemInfo, io.ReadCloser, transformedResourceMetadata, bool) {
	c.RLock()
	defer c.RUnlock()

	var meta transformedResourceMetadata
	filenameMeta, filenameContent := c.getFilenames(key)

	_, jsonContent, _ := c.FileCache.GetBytes(filenameMeta)
	if jsonContent == nil {
		return filecache.ItemInfo{}, nil, meta, false
	}

	if err := json.Unmarshal(jsonContent, &meta); err != nil {
		return filecache.ItemInfo{}, nil, meta, false
	}

	fi, rc, _ := c.FileCache.Get(filenameContent)

	return fi, rc, meta, rc != nil
}

func (c *ResourceCache) getFilenames(key string) (string, string) {
	filenameMeta := key + ".json"
	filenameContent := key + ".Content"

	return filenameMeta, filenameContent
}
