package entity

import "github.com/gohugonet/hugoverse/pkg/maps"

func (p *Page) Section() string {
	return p.Page.Section()
}

func (p *Page) Params() maps.Params {
	return p.Page.Params()
}

func (p *Page) Resources() PageResources {
	return p.resources
}
