package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/contenthub"

type PageGroup struct {
	key   string
	pages contenthub.Pages
}

func (p *PageGroup) Key() string {
	return p.key
}

func (p *PageGroup) Pages() contenthub.Pages {
	return p.pages
}

func (p *PageGroup) Append(page contenthub.Page) contenthub.Pages {
	return append(p.pages, page)
}
