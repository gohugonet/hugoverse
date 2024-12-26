package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/resource"
)

const nsCss = "css"

func registerCss(res resource.Resource) {
	f := func() *TemplateFuncsNamespace {
		ctx, err := resource.New(res)
		if err != nil {
			// TODO(bep) no panic.
			panic(err)
		}

		ns := &TemplateFuncsNamespace{
			Name:    nsCss,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Sass,
			[]string{"toCSS"},
			[][2]string{},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
