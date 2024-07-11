package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
)

type Page struct {
	*Source
	*valueobject.Content

	kind string
}

type TaxonomyPage struct {
	*Page

	singular string
}

type TermPage struct {
	*TaxonomyPage

	term string
}

func newPage(source *Source, content *valueobject.Content) (*Page, error) {
	p := &Page{
		Source:  source,
		Content: content,
		kind:    valueobject.KindPage,
	}

	return p, nil
}

func (p *Page) IsBundled() bool {
	return p.File.BundleType.IsContentResource()
}

func newTaxonomy(source *Source, content *valueobject.Content, singular string) (*TaxonomyPage, error) {
	p := &TaxonomyPage{
		Page:     &Page{Source: source, Content: content, kind: valueobject.KindTaxonomy},
		singular: singular,
	}

	return p, nil
}

func newTerm(source *Source, content *valueobject.Content, singular string, term string) (*TermPage, error) {
	p := &TermPage{
		TaxonomyPage: &TaxonomyPage{
			Page:     &Page{Source: source, Content: content, kind: valueobject.KindTerm},
			singular: singular,
		},
		term: term,
	}

	return p, nil
}
