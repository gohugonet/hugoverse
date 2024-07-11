package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

type PageBuilder struct {
	LangSvc     contenthub.LangService
	TaxonomySvc contenthub.TaxonomyService
	TemplateSvc contenthub.Template

	source          *Source
	sourceByte      []byte
	sourceParseInfo *valueobject.SourceParseInfo

	kind     string
	singular string
	term     string

	fm       *valueobject.FrontMatter
	fmParser *valueobject.FrontMatterParser

	sc *valueobject.ShortcodeParser
	c  *valueobject.Content
}

func (b *PageBuilder) WithSource(source *Source) *PageBuilder {
	b.source = source
	return b
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

func (b *PageBuilder) build() (contenthub.Page, error) {
	switch b.kind {
	case valueobject.KindHome, valueobject.KindSection:
		fmt.Println("build section or home, but it will not happen in this case")
	case valueobject.KindPage:
		return b.buildPage()
	case valueobject.KindTaxonomy:
		return b.buildTaxonomy()
	case valueobject.KindTerm:
		return b.buildTerm()
	default:
		return nil, fmt.Errorf("unknown kind %q", b.kind)
	}

	return nil, nil
}

func (b *PageBuilder) buildPage() (*Page, error) {
	return newPage(b.source, b.c)
}

func (b *PageBuilder) buildTaxonomy() (*TaxonomyPage, error) {
	singular := b.TaxonomySvc.Singular(b.source.File.Path().Path())
	return newTaxonomy(b.source, b.c, singular)
}

func (b *PageBuilder) buildTerm() (*TermPage, error) {
	p := b.source.File.Path()
	singular := b.TaxonomySvc.Singular(p.Path())
	term := p.Unnormalized().BaseNameNoIdentifier()

	return newTerm(b.source, b.c, singular, term)
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
	b.c = valueobject.NewContent()

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

	return nil
}

func (b *PageBuilder) parseKind() error {
	path := b.source.File.Path()

	kind := b.fm.Kind
	if kind == "" {
		kind = valueobject.KindSection

		if path.Base() == "/" {
			kind = valueobject.KindHome
		} else if b.source.File.IsBranchBundle() {
			// A section, taxonomy or term.
			if !b.TaxonomySvc.IsZero(path.Path()) {
				// Either a taxonomy or a term.
				if b.TaxonomySvc.PluralTreeKey(path.Path()) == path.Path() {
					kind = valueobject.KindTaxonomy
				} else {
					kind = valueobject.KindTerm
				}
			}
		} else {
			kind = valueobject.KindPage
		}
	}

	b.kind = kind

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
			LangService: b.LangSvc,
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
		b.c.AddReplacement(valueobject.InternalSummaryDividerPre, it)

		return nil
	}

}
