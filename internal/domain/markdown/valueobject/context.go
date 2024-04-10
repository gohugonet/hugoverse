package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/markdown"

type Context struct {
	*BufWriter
	positions []int
	markdown.ContextData
}

func (ctx *Context) PushPos(n int) {
	ctx.positions = append(ctx.positions, n)
}

func (ctx *Context) PopPos() int {
	i := len(ctx.positions) - 1
	p := ctx.positions[i]
	ctx.positions = ctx.positions[:i]
	return p
}
