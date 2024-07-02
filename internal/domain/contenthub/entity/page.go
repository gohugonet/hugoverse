package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

type Page struct {
	*Source
	*FrontMatter
	*Path
	*Shortcodes
	*Content

	kind     string
	singular string
	term     string

	taxonomyService contenthub.TaxonomyService

	bundled bool // Set if this page is bundled inside another.
}

func newBundledPage(source *Source, langSer contenthub.LangService, taxSer contenthub.TaxonomyService, tmplSvc contenthub.Template) (*Page, error) {
	p, err := newPage(source, langSer, taxSer, tmplSvc)
	if err != nil {
		return nil, err
	}
	p.bundled = true
	return p, nil
}

func newPage(source *Source, langSer contenthub.LangService, taxSer contenthub.TaxonomyService, tmplSvc contenthub.Template) (*Page, error) {
	contentBytes, err := source.contentSource()
	if err != nil {
		return nil, err
	}

	p := &Page{
		Source: source,
		FrontMatter: &FrontMatter{
			Params:     maps.Params{},
			Customized: maps.Params{},

			langService: langSer,
		},
		Shortcodes: &Shortcodes{source: contentBytes, ordinal: 0, tmplSvc: tmplSvc, pid: pid},
		Content:    &Content{source: contentBytes},

		bundled: false,

		taxonomyService: taxSer,
	}

	p.Source.registerHandler(p.FrontMatter.frontMatterHandler,
		p.Content.summaryHandler, p.Content.bytesHandler,
		p.Shortcodes.shortcodeHandler)

	if err := p.Source.parse(); err != nil {
		return nil, err
	}

	if err := p.FrontMatter.parse(); err != nil {
		return nil, err
	}

	p.setupPagePath()
	p.setupLang()
	p.setupKind()

	return p, nil
}

func (p *Page) setupPagePath() {
	pi := paths.Parse(p.Source.fi.Component(), p.Source.fi.FileName())
	if p.FrontMatter.Path != "" {
		p.Path = newPathFromConfig(p.FrontMatter.Path, p.FrontMatter.Kind, pi)
	} else {
		p.Path = &Path{pathInfo: pi}
	}
}

func (p *Page) setupLang() {
	l, ok := p.FrontMatter.langService.GetSourceLang(p.Source.fi.Root())
	if ok {
		idx, err := p.FrontMatter.langService.GetLanguageIndex(l)
		if err != nil {
			panic(err)
		}
		p.Identity.Lang = l
		p.Identity.LangIdx = idx
	} else {
		panic(fmt.Sprintf("unknown lang %q", p.Source.fi.Root()))
	}
}

func (p *Page) setupKind() {
	p.kind = p.FrontMatter.Kind
	if p.FrontMatter.Kind == "" {
		p.Kind = valueobject.KindSection
		if p.Path.pathInfo.Base() == "/" {
			p.Kind = valueobject.KindHome
		} else if p.Path.pathInfo.IsBranchBundle() {
			// A section, taxonomy or term.
			if !p.taxonomyService.IsZero(p.Path.Path()) {
				// Either a taxonomy or a term.
				if p.taxonomyService.PluralTreeKey(p.Path.Path()) == p.Path.Path() {
					p.Kind = valueobject.KindTaxonomy
					p.singular = p.taxonomyService.Singular(p.Path.Path())
				} else {
					p.Kind = valueobject.KindTerm
					p.singular = p.taxonomyService.Singular(p.Path.Path())
					p.term = p.Path.pathInfo.Unnormalized().BaseNameNoIdentifier()
				}
			}
		} else {
			p.Kind = valueobject.KindPage
		}
	}
}
