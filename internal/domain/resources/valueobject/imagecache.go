package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
)

type ImageCache struct {
	Fcache *filecache.Cache
	Mcache *dynacache.Partition[string, *ResourceAdapter]
}

type ResourceAdapter struct{}

func NewImageCache(fileCache *filecache.Cache, memCache *dynacache.Cache) *ImageCache {
	return &ImageCache{
		Fcache: fileCache,
		Mcache: dynacache.GetOrCreatePartition[string, *ResourceAdapter](
			memCache,
			"/imgs",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 70},
		),
	}
}
