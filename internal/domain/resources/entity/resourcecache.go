package entity

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"io"
	"path"
	"path/filepath"
	"strings"
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

func (c *ResourceCache) GetFromFile(key string) (filecache.ItemInfo, io.ReadCloser, valueobject.TransformedResourceMetadata, bool) {
	c.RLock()
	defer c.RUnlock()

	var meta valueobject.TransformedResourceMetadata
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

func (c *ResourceCache) CleanKey(key string) string {
	return strings.TrimPrefix(path.Clean(strings.ToLower(filepath.ToSlash(key))), "/")
}

// WriteMeta writes the metadata to file and returns a writer for the content part.
func (c *ResourceCache) WriteMeta(key string, meta valueobject.TransformedResourceMetadata) (filecache.ItemInfo, io.WriteCloser, error) {
	filenameMeta, filenameContent := c.getFilenames(key)
	raw, err := json.Marshal(meta)
	if err != nil {
		return filecache.ItemInfo{}, nil, err
	}

	_, fm, err := c.FileCache.WriteCloser(filenameMeta)
	if err != nil {
		return filecache.ItemInfo{}, nil, err
	}
	defer fm.Close()

	if _, err := fm.Write(raw); err != nil {
		return filecache.ItemInfo{}, nil, err
	}

	fi, fc, err := c.FileCache.WriteCloser(filenameContent)

	return fi, fc, err
}
