package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/template"

func IdentityOr(a, b template.Identity) template.Identity {
	return orIdentity{a: a, b: b}
}

type orIdentity struct {
	a, b template.Identity
}

func (o orIdentity) IdentifierBase() string {
	return o.a.IdentifierBase()
}
