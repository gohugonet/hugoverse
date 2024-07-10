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

	fm       *valueobject.FrontMatter
	fmParser *valueobject.FrontMatterParser

	sc *valueobject.ShortcodeParser
	c  *valueobject.Content
}

func (b *PageBuilder) WithSource(source *Source) {
	b.source = source
}

func (b *PageBuilder) Build() (*Page, error) {
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

	// new page based on those basic info

	return nil, nil
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
