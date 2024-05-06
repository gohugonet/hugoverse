package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"strings"
)

// FileCaches is a named set of caches.
type FileCaches map[string]*filecache.Cache

// Get gets a named cache, nil if none found.
func (f FileCaches) Get(name string) *filecache.Cache {
	return f[strings.ToLower(name)]
}

// GetJSONCache gets the file cache for getJSON.
func (f FileCaches) GetJSONCache() *filecache.Cache {
	return f[resources.KeyGetJSON]
}

// GetCSVCache gets the file cache for getCSV.
func (f FileCaches) GetCSVCache() *filecache.Cache {
	return f[resources.KeyGetCSV]
}

func (f FileCaches) ImageCache() *filecache.Cache {
	return f[resources.KeyImages]
}

// GetResourceCache gets the file cache for remote resources.
func (f FileCaches) GetResourceCache() *filecache.Cache {
	return f[resources.KeyGetResource]
}

// AssetsCache gets the file cache for assets (processed resources, SCSS etc.).
func (f FileCaches) AssetsCache() *filecache.Cache {
	return f[resources.KeyAssets]
}
