package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/hugo"
)

const nsHugo = "hugo"

func registerHugo(ver hugo.Version) {
	f := func() *TemplateFuncsNamespace {

		ns := &TemplateFuncsNamespace{
			Name:    nsHugo,
			Context: func(cctx context.Context, args ...any) (any, error) { return ver, nil },
		}

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
