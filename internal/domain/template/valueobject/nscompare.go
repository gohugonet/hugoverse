package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/compare"
)

const nsCompare = "compare"

func registerCompare(timezone compare.TimeZone) {
	f := func() *TemplateFuncsNamespace {
		ctx := compare.New(timezone.Location(), false)

		ns := &TemplateFuncsNamespace{
			Name:    nsCompare,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Default,
			[]string{"default"},
			[][2]string{
				{`{{ "Hugo Rocks!" | default "Hugo Rules!" }}`, `Hugo Rocks!`},
				{`{{ "" | default "Hugo Rules!" }}`, `Hugo Rules!`},
			},
		)

		ns.AddMethodMapping(ctx.Eq,
			[]string{"eq"},
			[][2]string{
				{`{{ if eq .Section "blog" }}current-section{{ end }}`, `current-section`},
			},
		)

		ns.AddMethodMapping(ctx.Ge,
			[]string{"ge"},
			[][2]string{
				{`{{ if ge hugo.Version "0.80" }}Reasonable new Hugo version!{{ end }}`, `Reasonable new Hugo version!`},
			},
		)

		ns.AddMethodMapping(ctx.Gt,
			[]string{"gt"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Le,
			[]string{"le"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Lt,
			[]string{"lt"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Ne,
			[]string{"ne"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.Conditional,
			[]string{"cond"},
			[][2]string{
				{`{{ cond (eq (add 2 2) 4) "2+2 is 4" "what?" | safeHTML }}`, `2+2 is 4`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
