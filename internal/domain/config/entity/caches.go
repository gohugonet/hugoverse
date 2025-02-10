package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/config/valueobject"
	"time"
)

type Caches struct {
	valueobject.CachesConfig
}

func (c Caches) CachesIterator(cb func(cacheKey string, isResourceDir bool, dir string, age time.Duration)) {
	for k, v := range c.CachesConfig {
		cb(k, v.IsResourceDir, v.DirCompiled, v.MaxAge)
	}
}
