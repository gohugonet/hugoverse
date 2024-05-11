package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/crypto"
)

const nsCrypto = "crypto"

func registerCrypto() {
	f := func() *TemplateFuncsNamespace {
		ctx := crypto.New()

		ns := &TemplateFuncsNamespace{
			Name:    nsCrypto,
			Context: func(cctx context.Context, args ...any) (any, error) { return ctx, nil },
		}

		ns.AddMethodMapping(ctx.MD5,
			[]string{"md5"},
			[][2]string{
				{`{{ md5 "Hello world, gophers!" }}`, `b3029f756f98f79e7f1b7f1d1f0dd53b`},
				{`{{ crypto.MD5 "Hello world, gophers!" }}`, `b3029f756f98f79e7f1b7f1d1f0dd53b`},
			},
		)

		ns.AddMethodMapping(ctx.SHA1,
			[]string{"sha1"},
			[][2]string{
				{`{{ sha1 "Hello world, gophers!" }}`, `c8b5b0e33d408246e30f53e32b8f7627a7a649d4`},
			},
		)

		ns.AddMethodMapping(ctx.SHA256,
			[]string{"sha256"},
			[][2]string{
				{`{{ sha256 "Hello world, gophers!" }}`, `6ec43b78da9669f50e4e422575c54bf87536954ccd58280219c393f2ce352b46`},
			},
		)

		ns.AddMethodMapping(ctx.FNV32a,
			nil,
			[][2]string{
				{`{{ crypto.FNV32a "Hugo Rocks!!" }}`, `1515779328`},
			},
		)

		ns.AddMethodMapping(ctx.HMAC,
			[]string{"hmac"},
			[][2]string{
				{`{{ hmac "sha256" "Secret key" "Hello world, gophers!" }}`, `b6d11b6c53830b9d87036272ca9fe9d19306b8f9d8aa07b15da27d89e6e34f40`},
			},
		)

		return ns
	}

	AddTemplateFuncsNamespace(f)
}
