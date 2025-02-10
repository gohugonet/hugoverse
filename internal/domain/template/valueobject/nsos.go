package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/os"
)

const nsOs = "os"

func registerOs(ws os.Os) {
	f := func() *TemplateFuncsNamespace {
		ctx := os.New(ws)

		ns := &TemplateFuncsNamespace{
			Name:    nsOs,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Getenv,
			[]string{"getenv"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.ReadDir,
			[]string{"readDir"},
			[][2]string{
				{`{{ range (readDir "files") }}{{ .Name }}{{ end }}`, "README.txt"},
			},
		)

		ns.AddMethodMapping(ctx.ReadFile,
			[]string{"readFile"},
			[][2]string{
				{`{{ readFile "files/README.txt" }}`, `Hugo Rocks!`},
			},
		)

		ns.AddMethodMapping(ctx.FileExists,
			[]string{"fileExists"},
			[][2]string{
				{`{{ fileExists "foo.txt" }}`, `false`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
