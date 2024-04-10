package valueobject

import "github.com/gohugonet/hugoverse/pkg/types/hstring"

type headingContext struct {
	page      any
	level     int
	anchor    string
	text      hstring.RenderedString
	plainText string
	*AttributesHolder
}

func (ctx headingContext) Page() any {
	return ctx.page
}

func (ctx headingContext) Level() int {
	return ctx.level
}

func (ctx headingContext) Anchor() string {
	return ctx.anchor
}

func (ctx headingContext) Text() hstring.RenderedString {
	return ctx.text
}

func (ctx headingContext) PlainText() string {
	return ctx.plainText
}
