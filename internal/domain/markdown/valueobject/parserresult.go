package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/markdown"

type ParserResult struct {
	doc any
	toc markdown.TocFragments
}

func NewParserResult(doc any, toc markdown.TocFragments) *ParserResult {
	return &ParserResult{
		doc: doc,
		toc: toc,
	}
}

func (p ParserResult) Doc() any {
	return p.doc
}

func (p ParserResult) TableOfContents() markdown.TocFragments {
	return p.toc
}
