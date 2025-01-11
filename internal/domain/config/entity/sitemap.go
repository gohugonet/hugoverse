package entity

import "github.com/gohugonet/hugoverse/internal/domain/config/valueobject"

type Sitemap struct {
	Conf valueobject.SitemapConfig
}

func (s Sitemap) ChangeFreq() string {
	return s.Conf.ChangeFreq
}

func (s Sitemap) Priority() float64 {
	return s.Conf.Priority
}
