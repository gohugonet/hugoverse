package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"sync"
)

type ResourceCache struct {
	sync.RWMutex

	cacheResource               *dynacache.Partition[string, resources.Resource]
	cacheResources              *dynacache.Partition[string, []resources.Resource]
	cacheResourceTransformation *dynacache.Partition[string, *resourceAdapterInner]

	fileCache *filecache.Cache
}
