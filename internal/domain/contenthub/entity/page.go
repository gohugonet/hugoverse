package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"time"
)

type Page struct {
	*Source
	*Content
	*Meta

	*Layout
	*Output
	contenthub.PagerManager

	store *maps.Scratch

	title   string
	kind    string
	pageMap *PageMap
}

func (p *Page) PageOutputs() ([]contenthub.PageOutput, error) {
	return p.Outputs(p)
}

func (p *Page) Layouts() []string {
	//TODO: multiple outputs, not supported yet
	// output map layout

	switch p.kind {
	case valueobject.KindHome:
		return p.Layout.home()
	case valueobject.KindPage:
		return p.Layout.page(p.source.File.Section(), p.source.File.BaseFileName())
	case valueobject.KindSection:
		return p.Layout.section(p.source.File.Section())
	case valueobject.KindTaxonomy:
		return p.Layout.taxonomy()
	case valueobject.KindTerm:
		return p.Layout.term()
	case valueobject.KindStatus404:
		return p.Layout.standalone404()
	case valueobject.KindSitemap:
		return p.Layout.standaloneSitemap()
	default:
		return nil
	}
}

type TaxonomyPage struct {
	*Page

	singular string
}

func (m *TaxonomyPage) Params() maps.Params {
	params := m.Page.Meta.Parameters
	params["Singular"] = m.singular
	return params
}

type TermPage struct {
	*TaxonomyPage

	term string
}

func (m *TermPage) Params() maps.Params {
	params := m.TaxonomyPage.Params()
	params["Term"] = m.term
	return params
}

func newPage(source *Source, content *Content) (*Page, error) {
	p := &Page{
		Source:  source,
		Content: content,
		Meta: &Meta{
			List:       Always,
			Parameters: map[string]any{},
			Date:       time.Now(),
		},

		kind: valueobject.KindPage,

		Layout: &Layout{},
	}

	return p, nil
}

func (p *Page) IsHome() bool {
	return p.Kind() == valueobject.KindHome
}

func (p *Page) IsPage() bool {
	return p.Kind() == valueobject.KindPage
}

func (p *Page) IsSection() bool {
	return p.Kind() == valueobject.KindSection
}

func (p *Page) Kind() string {
	return p.kind
}

func (p *Page) IsBundled() bool {
	return p.File.BundleType.IsContentResource()
}

func (p *Page) Eq(other contenthub.Page) bool {
	return p.Source.Identity.IdentifierBase() == other.PageIdentity().IdentifierBase()
}

func (p *Page) isStandalone() bool {
	res := false
	switch p.Kind() {
	case valueobject.KindStatus404, valueobject.KindRobotsTXT, valueobject.KindSitemap:
		res = true
	}

	return res
}

func (p *Page) isVirtualPage() bool {
	return p.Content == nil
}

func newTaxonomy(source *Source, content *Content, singular string) (*TaxonomyPage, error) {
	p, err := newPage(source, content)
	if err != nil {
		return nil, err
	}

	p.kind = valueobject.KindTaxonomy
	taxonomy := &TaxonomyPage{
		Page:     p,
		singular: singular,
	}

	taxonomy.Page.title = singular

	return taxonomy, nil
}

func newTerm(source *Source, content *Content, singular string, term string) (*TermPage, error) {
	taxonomy, err := newTaxonomy(source, content, singular)
	if err != nil {
		return nil, err
	}

	taxonomy.Page.kind = valueobject.KindTerm
	tp := &TermPage{
		TaxonomyPage: taxonomy,
		term:         term,
	}

	tp.TaxonomyPage.Page.title = term

	return tp, nil
}
