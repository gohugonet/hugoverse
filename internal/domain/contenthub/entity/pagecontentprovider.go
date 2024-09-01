package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/output"
)

type ContentProvider struct {
	sourceKey string
	content   *Content

	cache *Cache
	f     output.Format

	contentToRender []byte

	converter contenthub.Converter
}

func (c *ContentProvider) cacheKey() string {
	return c.sourceKey + "/" + c.f.Name
}

func (c *ContentProvider) Content() (any, error) {
	v, err := c.cache.CacheContentRendered.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[valueobject.ContentSummary], error) {

		return nil, nil
	})
	if err != nil {
		return valueobject.ContentSummary{}, err
	}

	return v.Value, nil
}

func (c *ContentProvider) toc() (valueobject.ContentToC, error) {
	v, err := c.cache.ContentTableOfContents.GetOrCreate(c.cacheKey(), func(string) (*stale.Value[valueobject.ContentToC], error) {
		return nil, nil
	})
	if err != nil {
		return valueobject.ContentToC{}, err
	}

	return v.Value, nil
}
