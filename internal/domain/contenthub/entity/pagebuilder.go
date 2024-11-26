package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

type PageBuilder struct {
	LangSvc     contenthub.LangService
	TaxonomySvc contenthub.TaxonomyService
	TemplateSvc contenthub.Template
	MediaSvc    contenthub.MediaService
	PageMapper  *PageMap

	Taxonomy   *Taxonomy
	Term       *Term
	Section    *Section
	Standalone *Standalone

	ConvertProvider *ContentSpec

	source          *Source
	sourceByte      []byte
	sourceParseInfo *valueobject.SourceParseInfo

	kind     string
	singular string
	term     string

	fm       *valueobject.FrontMatter
	fmParser *valueobject.FrontMatterParser

	sc *valueobject.ShortcodeParser
	c  *Content

	Log loggers.Logger
}

func (b *PageBuilder) WithSource(source *Source) *PageBuilder {
	cloneBuilder := *b
	cloneBuilder.reset()

	cloneBuilder.source = source // source changed, need light copy
	return &cloneBuilder
}

func (b *PageBuilder) reset() {
	b.c = nil
	b.kind = ""
}

func (b *PageBuilder) Build() (contenthub.Page, error) {
	if b.source == nil {
		return nil, fmt.Errorf("source for page builder is nil")
	}

	contentBytes, err := b.source.contentSource()
	if err != nil {
		return nil, err
	}

	b.sourceByte = contentBytes
	if err := b.parse(contentBytes); err != nil {
		return nil, err
	}

	return b.build()
}

func (b *PageBuilder) KindBuild() (contenthub.Page, error) {
	if b.source == nil {
		return nil, fmt.Errorf("source for page builder is nil")
	}

	if err := b.parseKind(); err != nil {
		return nil, err
	}

	if err := b.parseLanguageByDefault(); err != nil {
		return nil, err
	}

	b.fm = &valueobject.FrontMatter{}

	return b.build()
}

func (b *PageBuilder) build() (contenthub.Page, error) {
	switch b.kind {
	case valueobject.KindHome:
		return b.buildHome()
	case valueobject.KindSection:
		return b.buildSection()
	case valueobject.KindPage:
		return b.buildPage()
	case valueobject.KindTaxonomy:
		return b.buildTaxonomy()
	case valueobject.KindTerm:
		return b.buildTerm()
	case valueobject.KindStatus404:
		return b.build404()
	case valueobject.KindSitemap:
		return b.buildSitemap()
	default:
		return nil, fmt.Errorf("unknown kind %q", b.kind)
	}
}

func (b *PageBuilder) buildOutput(p *Page) error {
	p.Output = &Output{
		source:   p.Source,
		pageKind: p.Kind(),

		log: loggers.NewDefault(),
	}
	if err := p.Output.Build(b.ConvertProvider, b.TemplateSvc, b.MediaSvc); err != nil {
		return err
	}

	return nil
}

func (b *PageBuilder) buildPage() (*Page, error) {
	p, err := newPage(b.source, b.c)
	if err != nil {
		return nil, err
	}

	if err := b.applyFrontMatter(p); err != nil {
		return nil, err
	}
	p.pageMap = b.PageMapper
	if err := b.buildOutput(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (b *PageBuilder) buildPageWithKind(kind string) (*Page, error) {
	p, err := newPage(b.source, b.c)
	if err != nil {
		return nil, err
	}

	if err := b.applyFrontMatter(p); err != nil {
		return nil, err
	}
	p.pageMap = b.PageMapper
	p.kind = kind
	if p.kind == valueobject.KindSitemap || p.kind == valueobject.KindStatus404 {
		p.Meta.List = Never
	}
	if err := b.buildOutput(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (b *PageBuilder) applyFrontMatter(p *Page) error {
	p.title = b.fm.Title
	p.Meta.Weight = b.fm.Weight
	p.Meta.Parameters = b.fm.Params

	return nil
}

func (b *PageBuilder) buildHome() (*Page, error) {
	return b.buildPageWithKind(valueobject.KindHome)
}

func (b *PageBuilder) buildSection() (*Page, error) {
	return b.buildPageWithKind(valueobject.KindSection)
}

func (b *PageBuilder) build404() (*Page, error) {
	return b.buildPageWithKind(valueobject.KindStatus404)
}

func (b *PageBuilder) buildSitemap() (*Page, error) {
	return b.buildPageWithKind(valueobject.KindSitemap)
}

func (b *PageBuilder) buildTaxonomy() (*TaxonomyPage, error) {
	singular := b.Taxonomy.getTaxonomy(b.source.File.Paths().Path()).Singular()
	tp, err := newTaxonomy(b.source, b.c, singular)
	if err != nil {
		return nil, err
	}

	tp.pageMap = b.PageMapper

	if err := b.buildOutput(tp.Page); err != nil {
		return nil, err
	}

	return tp, nil
}

func (b *PageBuilder) buildTerm() (*TermPage, error) {
	p := b.source.File.Paths()
	singular := b.Taxonomy.getTaxonomy(p.Path()).Singular()
	term := p.Unnormalized().BaseNameNoIdentifier()

	t, err := newTerm(b.source, b.c, singular, term)
	if err != nil {
		return nil, err
	}

	t.pageMap = b.PageMapper

	if err := b.buildOutput(t.Page); err != nil {
		return nil, err
	}

	return t, nil
}

func (b *PageBuilder) parse(contentBytes []byte) error {
	pif, err := valueobject.NewSourceParseInfo(contentBytes, b)
	if err != nil {
		return err
	}

	if err := pif.Parse(); err != nil {
		return err
	}

	b.sourceParseInfo = pif
	b.sc = valueobject.NewShortcodeParser(contentBytes, b.source.Id, b.TemplateSvc)
	b.c = NewContent(contentBytes)

	if err := pif.Handle(); err != nil {
		return err
	}

	if err := b.parseFrontMatter(); err != nil {
		return err
	}
	if err := b.parseLanguage(); err != nil {
		return err
	}
	if err := b.parseKind(); err != nil {
		return err
	}
	if err := b.parseTerms(); err != nil {
		return err
	}

	return nil
}

func (b *PageBuilder) parseTerms() error {
	b.Term.Terms = b.fm.Terms
	return nil
}

func (b *PageBuilder) parseKind() error {
	path := b.source.File.Paths()

	kind := ""
	if b.fm != nil {
		kind = b.fm.Kind
	}

	if kind == "" {
		kind = valueobject.KindPage
		base := path.BaseNoLeadingSlash()

		switch base {
		case PageHomeBase, "":
			kind = valueobject.KindHome
		case StandalonePage404Base:
			kind = valueobject.KindStatus404
		case StandalonePageSitemapBase:
			kind = valueobject.KindSitemap
		default:
			if b.source.File.IsBranchBundle() {
				// A section, taxonomy or term.
				kind = valueobject.KindSection

				v := b.Taxonomy.getTaxonomy(path.Path())
				if !b.Taxonomy.IsZero(v) {
					// Either a taxonomy or a term.
					if b.Taxonomy.IsTaxonomyPath(path.Path()) {
						kind = valueobject.KindTaxonomy
					} else {
						kind = valueobject.KindTerm
					}
				}
			}
		}
	}

	b.kind = kind

	return nil
}

func (b *PageBuilder) parseLanguageByDefault() error {
	dl := b.LangSvc.DefaultLanguage()
	idx, err := b.LangSvc.GetLanguageIndex(dl)
	if err != nil {
		return fmt.Errorf("failed to get language index for %q: %s", dl, err)
	}

	b.source.Identity.Lang = dl
	b.source.Identity.LangIdx = idx
	return nil
}

func (b *PageBuilder) parseLanguage() error {
	l, ok := b.LangSvc.GetSourceLang(b.source.File.Root())
	if ok {
		idx, err := b.LangSvc.GetLanguageIndex(l)
		if err != nil {
			return fmt.Errorf("failed to get language index for %q: %s", l, err)
		}
		b.source.Identity.Lang = l
		b.source.Identity.LangIdx = idx
	} else {
		return fmt.Errorf("unknown lang %q", b.source.File.Root())
	}

	return nil
}

func (b *PageBuilder) parseFrontMatter() error {
	if b.fmParser == nil {
		return fmt.Errorf("front matter parser is nil")
	}
	fm, err := b.fmParser.Parse()
	if err != nil {
		return err
	}
	b.fm = fm

	return nil
}

func (b *PageBuilder) FrontMatterHandler() valueobject.ItemHandler {
	return func(it pageparser.Item) error {
		f := pageparser.FormatFromFrontMatterType(it.Type)

		m, err := metadecoders.Default.UnmarshalToMap(it.Val(b.sourceByte), f)

		if err != nil {
			return err
		}

		maps.PrepareParams(m)

		b.fmParser = &valueobject.FrontMatterParser{
			Params:      m,
			LangSvc:     b.LangSvc,
			TaxonomySvc: b.TaxonomySvc,
		}

		return nil
	}
}

func (b *PageBuilder) ShortcodeHandler() valueobject.IterHandler {
	return func(it pageparser.Item, pt *pageparser.Iterator) error {
		currShortcode, err := b.sc.ParseItem(it, pt)
		if err != nil {
			return err
		}

		b.c.AddShortcode(currShortcode)
		return nil
	}
}

func (b *PageBuilder) BytesHandler() valueobject.ItemHandler {
	return func(item pageparser.Item) error {
		b.c.AddItems(item)
		return nil
	}
}

func (b *PageBuilder) SummaryHandler() valueobject.IterHandler {
	return func(it pageparser.Item, iter *pageparser.Iterator) error {
		posBody := -1
		f := func(item pageparser.Item) bool {
			if posBody == -1 && !item.IsDone() {
				posBody = item.Pos()
			}

			if item.IsNonWhitespace(b.sourceByte) {
				b.c.SetSummaryTruncated()

				// Done
				return false
			}
			return true
		}
		iter.PeekWalk(f)

		b.c.SetSummaryDivider()

		// The content may be rendered by Goldmark or similar,
		// and we need to track the summary.
		b.c.AddReplacement(InternalSummaryDividerPre, it)

		return nil
	}

}
