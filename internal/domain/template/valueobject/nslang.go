package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/lang"
)

const nsLang = "lang"

func registerLang() {
	f := func() *TemplateFuncsNamespace {
		ctx := lang.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsLang,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.Translate,
			[]string{"i18n", "T"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.FormatNumber,
			nil,
			[][2]string{
				{`{{ 512.5032 | lang.FormatNumber 2 }}`, `512.50`},
			},
		)

		ns.AddMethodMapping(ctx.FormatPercent,
			nil,
			[][2]string{
				{`{{ 512.5032 | lang.FormatPercent 2 }}`, `512.50%`},
			},
		)

		ns.AddMethodMapping(ctx.FormatCurrency,
			nil,
			[][2]string{
				{`{{ 512.5032 | lang.FormatCurrency 2 "USD" }}`, `$512.50`},
			},
		)

		ns.AddMethodMapping(ctx.FormatAccounting,
			nil,
			[][2]string{
				{`{{ 512.5032 | lang.FormatAccounting 2 "NOK" }}`, `NOK512.50`},
			},
		)

		ns.AddMethodMapping(ctx.FormatNumberCustom,
			nil,
			[][2]string{
				{`{{ lang.FormatNumberCustom 2 12345.6789 }}`, `12,345.68`},
				{`{{ lang.FormatNumberCustom 2 12345.6789 "- , ." }}`, `12.345,68`},
				{`{{ lang.FormatNumberCustom 6 -12345.6789 "- ." }}`, `-12345.678900`},
				{`{{ lang.FormatNumberCustom 0 -12345.6789 "- . ," }}`, `-12,346`},
				{`{{ lang.FormatNumberCustom 0 -12345.6789 "-|.| " "|" }}`, `-12 346`},
				{`{{ -98765.4321 | lang.FormatNumberCustom 2 }}`, `-98,765.43`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
