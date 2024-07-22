package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/lazy"
)

type Page struct {
	*Source

	*valueobject.Content
	*Output

	kind string

	// Lazily initialized dependencies.
	lazyInit *lazy.Init
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

	if err := p.outputSetup(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Page) Kind() string {
	return p.kind
}

func (p *Page) IsBundled() bool {
	return p.File.BundleType.IsContentResource()
}

func (p *Page) isStandalone() bool {
	res := false
	switch p.Kind() {
	case valueobject.KindStatus404, valueobject.KindRobotsTXT, valueobject.KindSitemap:
		res = true
	}

	return res
}

func (p *Page) outputSetup() error {
	p.lazyInit.Add(func() (any, error) {
		p.Output = &Output{
			source:   p.Source,
			pageKind: p.Kind(),
		}
		if err := p.Output.Build(); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return nil
}

func newTaxonomy(source *Source, content *valueobject.Content, singular string) (*TaxonomyPage, error) {
	p, err := newPage(source, content)
	if err != nil {
		return nil, err
	}

	p.kind = valueobject.KindTaxonomy
	taxonomy := &TaxonomyPage{
		Page:     p,
		singular: singular,
	}

	return taxonomy, nil
}

func newTerm(source *Source, content *valueobject.Content, singular string, term string) (*TermPage, error) {
	taxonomy, err := newTaxonomy(source, content, singular)
	if err != nil {
		return nil, err
	}

	taxonomy.Page.kind = valueobject.KindTerm
	tp := &TermPage{
		TaxonomyPage: taxonomy,
		term:         term,
	}

	return tp, nil
}
