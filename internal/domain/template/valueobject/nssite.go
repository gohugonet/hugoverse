package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/site"
)

const nsSite = "site"

func registerSite(svc site.Service) {
	f := func() *TemplateFuncsNamespace {
		s := site.New(svc)

		ns := &TemplateFuncsNamespace{
			Name: nsSite,
			Context: func(cctx context.Context, args ...any) (any, error) {
				return s, nil
			},
		}

		// We just add the Site as the namespace here. No method mappings.

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
