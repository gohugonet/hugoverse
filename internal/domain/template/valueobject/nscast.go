package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/cast"
)

func registerCast() {
	f := func() *TemplateFuncsNamespace {
		ctx := cast.New()

		ns := &TemplateFuncsNamespace{
			Name:    "cast",
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.ToInt,
			[]string{"int"},
			[][2]string{
				{`{{ "1234" | int | printf "%T" }}`, `int`},
			},
		)

		ns.AddMethodMapping(ctx.ToString,
			[]string{"string"},
			[][2]string{
				{`{{ 1234 | string | printf "%T" }}`, `string`},
			},
		)

		ns.AddMethodMapping(ctx.ToFloat,
			[]string{"float"},
			[][2]string{
				{`{{ "1234" | float | printf "%T" }}`, `float64`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
