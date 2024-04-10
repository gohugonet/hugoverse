package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"sync"
)

type pageMeta struct {
	kind string

	title string

	markup      string
	contentType string

	layout string

	// Set if this page is bundled inside another.
	bundled bool

	sections []string

	contentConverterInit  sync.Once
	contentConverter      contenthub.Converter
	contentCovertProvider contenthub.ContentConvertProvider

	f contenthub.File
}

func (p *pageMeta) setMetadata() {
	p.markup = "markdown"
}

func (p *pageMeta) applyDefaultValues() { // buildConfig, markup, title
	if p.markup == "" {
		p.markup = "markdown"
	}

	p.title = "hardcode title"
}

func (p *pageMeta) File() contenthub.File {
	return p.f
}

func (p *pageMeta) Kind() string {
	return p.kind
}

func (p *pageMeta) SectionsEntries() []string {
	return p.sections
}

const defaultContentType = "page"

func (p *pageMeta) Type() string {
	return defaultContentType
}

func (p *pageMeta) Layout() string {
	return p.layout
}

func (p *pageMeta) noLink() bool {
	return false
}

func (p *pageMeta) newContentConverter(ps *pageState, markup string) (contenthub.Converter, error) {
	if ps == nil {
		panic("no Page provided")
	}
	cp := p.contentCovertProvider.GetContentConvertProvider(markup)
	if cp == nil {
		panic(fmt.Errorf("no content renderer found for markup %q", p.markup))
	}

	var id string
	var filename string
	var path string
	if !p.f.IsZero() {
		id = p.f.UniqueID()
		filename = p.f.Filename()
		path = p.f.Path()
	} else {
		panic("no file provided")
	}

	cpp, err := cp.New(
		markdown.DocumentContext{
			Document:     nil, //TODO
			DocumentID:   id,
			DocumentName: path,
			Filename:     filename,
		},
	)
	if err != nil {
		panic(err)
	}

	return cpp, nil
}
