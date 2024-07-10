package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
)

type Page struct {
	*Source
	*valueobject.FrontMatter
	*valueobject.ShortcodeParser
	*valueobject.Content

	kind     string
	singular string
	term     string

	taxonomyService contenthub.TaxonomyService
}

func newBundledPage(source *Source, langSer contenthub.LangService, taxSer contenthub.TaxonomyService, tmplSvc contenthub.Template) (*Page, error) {
	p, err := newPage(source, langSer, taxSer, tmplSvc)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func newPage(source *Source, langSer contenthub.LangService, taxSer contenthub.TaxonomyService, tmplSvc contenthub.Template) (*Page, error) {
	contentBytes, err := source.contentSource()
	if err != nil {
		return nil, err
	}

	p := &Page{
		Source: source,
		FrontMatter: &valueobject.FrontMatter{
			Params:     maps.Params{},
			Customized: maps.Params{},

			langService: langSer,
		},
		ShortcodeParser: &valueobject.ShortcodeParser{source: contentBytes, ordinal: 0, tmplSvc: tmplSvc, pid: source.Id},
		Content:         &valueobject.Content{source: contentBytes},

		taxonomyService: taxSer,
	}

	p.Source.registerHandler(p.FrontMatter.frontMatterHandler,
		p.Content.summaryHandler, p.Content.bytesHandler,
		p.ShortcodeParser.shortcodeHandler)

	if err := p.Source.parse(); err != nil {
		return nil, err
	}

	if err := p.FrontMatter.parse(); err != nil {
		return nil, err
	}

	p.setupLang()
	p.setupKind()

	return p, nil
}

func (p *Page) IsBundled() bool {
	return p.File.BundleType.IsContentResource()
}

func (p *Page) setupLang() {
	l, ok := p.FrontMatter.langService.GetSourceLang(p.Source.File.Root())
	if ok {
		idx, err := p.FrontMatter.langService.GetLanguageIndex(l)
		if err != nil {
			panic(err)
		}
		p.Identity.Lang = l
		p.Identity.LangIdx = idx
	} else {
		panic(fmt.Sprintf("unknown lang %q", p.Source.File.Root()))
	}
}

func (p *Page) setupKind() {
	path := p.File.Path()

	p.kind = p.FrontMatter.Kind
	if p.FrontMatter.Kind == "" {
		p.Kind = valueobject.KindSection
		if path.Base() == "/" {
			p.Kind = valueobject.KindHome
		} else if p.File.IsBranchBundle() {
			// A section, taxonomy or term.
			if !p.taxonomyService.IsZero(path.Path()) {
				// Either a taxonomy or a term.
				if p.taxonomyService.PluralTreeKey(path.Path()) == path.Path() {
					p.Kind = valueobject.KindTaxonomy
					p.singular = p.taxonomyService.Singular(path.Path())
				} else {
					p.Kind = valueobject.KindTerm
					p.singular = p.taxonomyService.Singular(path.Path())
					p.term = path.Unnormalized().BaseNameNoIdentifier()
				}
			}
		} else {
			p.Kind = valueobject.KindPage
		}
	}
}
