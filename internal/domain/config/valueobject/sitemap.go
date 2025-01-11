package valueobject

import "github.com/mitchellh/mapstructure"

type SitemapConfig struct {
	// The page change frequency.
	ChangeFreq string
	// The priority of the page.
	Priority float64
	// The sitemap filename.
	Filename string
	// Whether to disable page inclusion.
	Disable bool
}

func DecodeSitemap(prototype SitemapConfig, input map[string]any) (SitemapConfig, error) {
	err := mapstructure.WeakDecode(input, &prototype)
	return prototype, err
}
