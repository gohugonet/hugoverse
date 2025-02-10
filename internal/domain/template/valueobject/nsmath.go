package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/math"
)

const nsMath = "math"

func registerMath() {
	f := func() *TemplateFuncsNamespace {
		ctx := math.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsMath,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Abs,
			nil,
			[][2]string{
				{"{{ math.Abs -2.1 }}", "2.1"},
			},
		)

		ns.AddMethodMapping(ctx.Add,
			[]string{"add"},
			[][2]string{
				{"{{ add 1 2 }}", "3"},
			},
		)

		ns.AddMethodMapping(ctx.Ceil,
			nil,
			[][2]string{
				{"{{ math.Ceil 2.1 }}", "3"},
			},
		)

		ns.AddMethodMapping(ctx.Div,
			[]string{"div"},
			[][2]string{
				{"{{ div 6 3 }}", "2"},
			},
		)

		ns.AddMethodMapping(ctx.Floor,
			nil,
			[][2]string{
				{"{{ math.Floor 1.9 }}", "1"},
			},
		)

		ns.AddMethodMapping(ctx.Log,
			nil,
			[][2]string{
				{"{{ math.Log 1 }}", "0"},
			},
		)

		ns.AddMethodMapping(ctx.Max,
			nil,
			[][2]string{
				{"{{ math.Max 1 2 }}", "2"},
			},
		)

		ns.AddMethodMapping(ctx.Min,
			nil,
			[][2]string{
				{"{{ math.Min 1 2 }}", "1"},
			},
		)

		ns.AddMethodMapping(ctx.Mod,
			[]string{"mod"},
			[][2]string{
				{"{{ mod 15 3 }}", "0"},
			},
		)

		ns.AddMethodMapping(ctx.ModBool,
			[]string{"modBool"},
			[][2]string{
				{"{{ modBool 15 3 }}", "true"},
			},
		)

		ns.AddMethodMapping(ctx.Mul,
			[]string{"mul"},
			[][2]string{
				{"{{ mul 2 3 }}", "6"},
			},
		)

		ns.AddMethodMapping(ctx.Pow,
			[]string{"pow"},
			[][2]string{
				{"{{ math.Pow 2 3 }}", "8"},
			},
		)

		ns.AddMethodMapping(ctx.Rand,
			nil,
			[][2]string{
				{"{{ math.Rand }}", "0.6312770459590062"},
			},
		)

		ns.AddMethodMapping(ctx.Round,
			nil,
			[][2]string{
				{"{{ math.Round 1.5 }}", "2"},
			},
		)

		ns.AddMethodMapping(ctx.Sqrt,
			nil,
			[][2]string{
				{"{{ math.Sqrt 81 }}", "9"},
			},
		)

		ns.AddMethodMapping(ctx.Sub,
			[]string{"sub"},
			[][2]string{
				{"{{ sub 3 2 }}", "1"},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
