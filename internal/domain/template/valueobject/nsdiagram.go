package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/diagrams"
)

const nsDiagram = "diagrams"

func registerDiagram() {
	f := func() *TemplateFuncsNamespace {
		ctx := diagrams.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsDiagram,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Goat,
			[]string{"goat"},
			[][2]string{
				{`{{ Goat "reader" }}`, `svg`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
