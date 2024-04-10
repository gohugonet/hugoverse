package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/types/hstring"
)

type imageLinkContext struct {
	linkContext
	ordinal int
	isBlock bool
}

func (ctx imageLinkContext) IsBlock() bool {
	return ctx.isBlock
}

func (ctx imageLinkContext) Ordinal() int {
	return ctx.ordinal
}

type linkContext struct {
	page        any
	destination string
	title       string
	text        hstring.RenderedString
	plainText   string
	*AttributesHolder
}

func (ctx linkContext) Destination() string {
	return ctx.destination
}

func (ctx linkContext) Page() any {
	return ctx.page
}

func (ctx linkContext) Text() hstring.RenderedString {
	return ctx.text
}

func (ctx linkContext) PlainText() string {
	return ctx.plainText
}

func (ctx linkContext) Title() string {
	return ctx.title
}
