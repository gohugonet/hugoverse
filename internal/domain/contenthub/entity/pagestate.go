package entity

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

type pageState struct {
	// This slice will be of same length as the number of global slice of output
	// formats (for all sites).
	pageOutputs []*pageOutput

	// This will be shifted out when we start to render a new output format.
	*pageOutput

	// Common for all output formats.
	*pageCommon
}

func (p *pageState) mapContent(meta *pageMeta) error {
	p.cmap = &pageContentMap{
		items: make([]any, 0, 20),
	}
	return p.mapContentForResult(
		p.source.parsed,
		p.cmap,
		meta.markup,
	)
}

func (p *pageState) mapContentForResult(result pageparser.Result, rn *pageContentMap, markup string) error {
	iter := result.Iterator()
	fail := func(err error, i pageparser.Item) error {
		return errors.New("fail fail fail")
	}

Loop:
	for {
		it := iter.Next()

		switch {
		case it.Type == pageparser.TypeIgnore:
		case it.IsFrontMatter():
			panic("not implemented front matter yet")
		case it.Type == pageparser.TypeLeadSummaryDivider:
			panic("not implemented lead summary divider yet")
		case it.Type == pageparser.TypeEmoji:
			panic("not implemented emoji yet")
		case it.IsEOF():
			break Loop
		case it.IsError():
			err := fail(errors.New(it.ValStr(result.Input())), it)
			return err

		default:
			rn.AddBytes(it)
		}
	}

	return nil
}

// This is serialized
func (p *pageState) initOutputFormat() error {
	if err := p.shiftToOutputFormat(); err != nil {
		return err
	}

	return nil
}

// shiftToOutputFormat is serialized. The output format idx refers to the
// full set of output formats for all sites.
func (p *pageState) shiftToOutputFormat() error {
	if err := p.initPage(); err != nil {
		return err
	}

	p.pageOutput = p.pageOutputs[0]
	if p.pageOutput == nil {
		panic(fmt.Sprintf("pageOutput is nil for output idx %d", 0))
	}

	cp := p.pageOutput.cp
	if cp == nil {
		var err error
		cp, err = newPageContentOutput(p)
		if err != nil {
			return err
		}
	}
	p.pageOutput.initContentProvider(cp)

	return nil
}

// Must be run after the site section tree etc. is built and ready.
func (p *pageState) initPage() error {
	if _, err := p.init.Do(); err != nil {
		return err
	}
	return nil
}

func (p *pageState) getContentConverter() contenthub.Converter {
	var err error
	p.m.contentConverterInit.Do(func() {
		markup := p.m.markup
		if markup == "html" {
			// Only used for shortcode inner content.
			markup = "markdown"
		}
		p.m.contentConverter, err = p.m.newContentConverter(p, markup)
	})

	if err != nil {
		fmt.Printf("Failed to create content converter: %v", err)
	}
	return p.m.contentConverter
}

func (p *pageState) getLayoutDescriptor() valueobject.LayoutDescriptor {
	p.layoutDescriptorInit.Do(func() {
		var section string
		sections := p.SectionsEntries()

		switch p.Kind() {
		case contenthub.KindSection:
			if len(sections) > 0 {
				section = sections[0]
			}
		default:
		}

		p.layoutDescriptor = valueobject.LayoutDescriptor{
			Kind:    p.Kind(),
			Type:    p.Type(),
			Lang:    "en",
			Layout:  p.Layout(),
			Section: section,
		}
	})

	return p.layoutDescriptor
}
