package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/pkg/template/funcs/hugo"
)

const nsHugo = "hugo"

func registerHugo(info hugo.Info) {
	f := func() *TemplateFuncsNamespace {
		h := hugo.New(info)

		ns := &TemplateFuncsNamespace{
			Name: nsHugo,
			Context: func(cctx context.Context, args ...any) (any, error) {
				return h, nil
			},
		}

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
