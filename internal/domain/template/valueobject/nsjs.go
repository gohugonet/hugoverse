package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/js"
)

const nsJs = "js"

func registerJs(client js.Client) {
	f := func() *TemplateFuncsNamespace {
		ctx, err := js.New(client)
		if err != nil {
			// TODO(bep) no panic.
			panic(err)
		}

		ns := &TemplateFuncsNamespace{
			Name:    nsJs,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
