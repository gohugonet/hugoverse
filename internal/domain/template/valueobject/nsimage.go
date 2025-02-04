package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/image"
)

const nsImage = "images"

func registerImages(img image.Image) {
	f := func() *TemplateFuncsNamespace {
		ctx, err := image.New(img)
		if err != nil {
			// TODO(bep) no panic.
			panic(err)
		}

		ns := &TemplateFuncsNamespace{
			Name:    nsImage,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
