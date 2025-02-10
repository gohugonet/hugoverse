package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/fmt"
)

const nsFmt = "fmt"

func registerFmt() {
	f := func() *TemplateFuncsNamespace {
		ctx := fmt.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsFmt,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Print,
			[]string{"print"},
			[][2]string{
				{`{{ print "works!" }}`, `works!`},
			},
		)

		ns.AddMethodMapping(ctx.Println,
			[]string{"println"},
			[][2]string{
				{`{{ println "works!" }}`, "works!\n"},
			},
		)

		ns.AddMethodMapping(ctx.Printf,
			[]string{"printf"},
			[][2]string{
				{`{{ printf "%s!" "works" }}`, `works!`},
			},
		)

		ns.AddMethodMapping(ctx.Errorf,
			[]string{"errorf"},
			[][2]string{
				{`{{ errorf "%s." "failed" }}`, ``},
			},
		)

		ns.AddMethodMapping(ctx.Erroridf,
			[]string{"erroridf"},
			[][2]string{
				{`{{ erroridf "my-err-id" "%s." "failed" }}`, ``},
			},
		)

		ns.AddMethodMapping(ctx.Warnidf,
			[]string{"warnidf"},
			[][2]string{
				{`{{ warnidf "my-warn-id" "%s." "warning" }}`, ``},
			},
		)

		ns.AddMethodMapping(ctx.Warnf,
			[]string{"warnf"},
			[][2]string{
				{`{{ warnf "%s." "warning" }}`, ``},
			},
		)
		return ns
	}

	AddTemplateFuncsNamespace(f)
}
