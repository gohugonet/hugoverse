package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/inflect"
)

const nsInflect = "inflect"

func registerInflect() {
	f := func() *TemplateFuncsNamespace {
		ctx := inflect.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsInflect,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Humanize,
			[]string{"humanize"},
			[][2]string{
				{`{{ humanize "my-first-post" }}`, `My first post`},
				{`{{ humanize "myCamelPost" }}`, `My camel post`},
				{`{{ humanize "52" }}`, `52nd`},
				{`{{ humanize 103 }}`, `103rd`},
			},
		)

		ns.AddMethodMapping(ctx.Pluralize,
			[]string{"pluralize"},
			[][2]string{
				{`{{ "cat" | pluralize }}`, `cats`},
			},
		)

		ns.AddMethodMapping(ctx.Singularize,
			[]string{"singularize"},
			[][2]string{
				{`{{ "cats" | singularize }}`, `cat`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
