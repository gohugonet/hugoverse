package valueobject

import (
	"context"
	"errors"
	translators "github.com/gohugoio/localescompressed"
	"github.com/mdfriday/hugoverse/pkg/htime"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/time"
	sysTime "time"
)

const nsTime = "time"

func registerTime() {
	f := func() *TemplateFuncsNamespace {
		trans := translators.GetTranslator("en") //TODO, make it more extensible
		formatter := htime.NewTimeFormatter(trans)

		ctx := time.New(formatter, sysTime.UTC)

		ns := &TemplateFuncsNamespace{
			Name: nsTime,
			Context: func(cctx context.Context, args ...any) (any, error) {
				// Handle overlapping "time" namespace and func.
				//
				// If no args are passed to `time`, assume namespace usage and
				// return namespace context.
				//
				// If args are passed, call AsTime().

				switch len(args) {
				case 0:
					return ctx, nil
				case 1:
					return ctx.AsTime(args[0])
				case 2:
					return ctx.AsTime(args[0], args[1])

				// 3 or more arguments. Currently not supported.
				default:
					return nil, errors.New("invalid arguments supplied to `time`")
				}
			},
		}

		ns.AddMethodMapping(ctx.Format,
			[]string{"dateFormat"},
			[][2]string{
				{`dateFormat: {{ dateFormat "Monday, Jan 2, 2006" "2015-01-21" }}`, `dateFormat: Wednesday, Jan 21, 2015`},
			},
		)

		ns.AddMethodMapping(ctx.Now,
			[]string{"now"},
			[][2]string{},
		)

		ns.AddMethodMapping(ctx.AsTime,
			nil,
			[][2]string{
				{`{{ (time "2015-01-21").Year }}`, `2015`},
			},
		)

		ns.AddMethodMapping(ctx.Duration,
			[]string{"duration"},
			[][2]string{
				{`{{ mul 60 60 | duration "second" }}`, `1h0m0s`},
			},
		)

		ns.AddMethodMapping(ctx.ParseDuration,
			nil,
			[][2]string{
				{`{{ "1h12m10s" | time.ParseDuration }}`, `1h12m10s`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
