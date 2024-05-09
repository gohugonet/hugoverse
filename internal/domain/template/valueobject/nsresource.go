package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/resource"
)

const nsResources = "resources"

func registerResources(res resource.Resource) {
	f := func() *TemplateFuncsNamespace {
		ctx, err := resource.New(res)
		if err != nil {
			// TODO(bep) no panic.
			panic(err)
		}

		ns := &TemplateFuncsNamespace{
			Name:    nsResources,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Get,
			nil,
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Minify,
			[]string{"minify"},
			[][2]string{},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
