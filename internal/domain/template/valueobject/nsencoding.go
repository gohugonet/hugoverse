package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/encoding"
)

const nsEncoding = "encoding"

func registerEncoding() {
	f := func() *TemplateFuncsNamespace {
		ctx := encoding.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsEncoding,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Base64Decode,
			[]string{"base64Decode"},
			[][2]string{
				{`{{ "SGVsbG8gd29ybGQ=" | base64Decode }}`, `Hello world`},
				{`{{ 42 | base64Encode | base64Decode }}`, `42`},
			},
		)

		ns.AddMethodMapping(ctx.Base64Encode,
			[]string{"base64Encode"},
			[][2]string{
				{`{{ "Hello world" | base64Encode }}`, `SGVsbG8gd29ybGQ=`},
			},
		)

		ns.AddMethodMapping(ctx.Jsonify,
			[]string{"jsonify"},
			[][2]string{
				{`{{ (slice "A" "B" "C") | jsonify }}`, `["A","B","C"]`},
				{`{{ (slice "A" "B" "C") | jsonify (dict "indent" "  ") }}`, "[\n  \"A\",\n  \"B\",\n  \"C\"\n]"},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
