package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin/valueobject"
)

type Cache struct {
	Conf *valueobject.Config
}

func (a *Cache) ETage() string      { return a.Conf.Etag }
func (a *Cache) NewETage() string   { return valueobject.NewEtag() }
func (a *Cache) CacheMaxAge() int64 { return a.Conf.CacheMaxAge }
