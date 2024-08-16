package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/reflect"
)

const nsReflect = "reflect"

func registerReflect() {
	f := func() *TemplateFuncsNamespace {
		ctx := reflect.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsReflect,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.IsMap,
			nil,
			[][2]string{
				{`{{ if reflect.IsMap (dict "a" 1) }}Map{{ end }}`, `Map`},
			},
		)

		ns.AddMethodMapping(ctx.IsSlice,
			nil,
			[][2]string{
				{`{{ if reflect.IsSlice (slice 1 2 3) }}Slice{{ end }}`, `Slice`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
