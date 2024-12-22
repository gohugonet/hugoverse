package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/urls"
)

const nsUrls = "urls"

func registerUrls(url urls.URL, ref urls.RefSource) {
	f := func() *TemplateFuncsNamespace {
		ctx := urls.New(url, ref)

		ns := &TemplateFuncsNamespace{
			Name:    nsUrls,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.AbsURL,
			[]string{"absURL"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.AbsLangURL,
			[]string{"absLangURL"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.Ref,
			[]string{"ref"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.RelURL,
			[]string{"relURL"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.RelLangURL,
			[]string{"relLangURL"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.RelRef,
			[]string{"relref"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.URLize,
			[]string{"urlize"},
			[][2]string{},
		)
		ns.AddMethodMapping(ctx.JoinPath,
			nil,
			[][2]string{
				{`{{ urls.JoinPath "https://example.org" "foo" }}`, `https://example.org/foo`},
				{`{{ urls.JoinPath (slice "a" "b") }}`, `a/b`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
