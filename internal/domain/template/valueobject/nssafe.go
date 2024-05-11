package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/safe"
)

const nsSafe = "safe"

func registerSafe() {
	f := func() *TemplateFuncsNamespace {
		ctx := safe.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsSafe,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.CSS,
			[]string{"safeCSS"},
			[][2]string{
				{`{{ "Bat&Man" | safeCSS | safeCSS }}`, `Bat&amp;Man`},
			},
		)

		ns.AddMethodMapping(ctx.HTML,
			[]string{"safeHTML"},
			[][2]string{
				{`{{ "Bat&Man" | safeHTML | safeHTML }}`, `Bat&Man`},
				{`{{ "Bat&Man" | safeHTML }}`, `Bat&Man`},
			},
		)

		ns.AddMethodMapping(ctx.HTMLAttr,
			[]string{"safeHTMLAttr"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.JS,
			[]string{"safeJS"},
			[][2]string{
				{`{{ "(1*2)" | safeJS | safeJS }}`, `(1*2)`},
			},
		)

		ns.AddMethodMapping(ctx.JSStr,
			[]string{"safeJSStr"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.URL,
			[]string{"safeURL"},
			[][2]string{
				{`{{ "http://gohugo.io" | safeURL | safeURL }}`, `http://gohugo.io`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
