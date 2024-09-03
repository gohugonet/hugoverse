package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/types"
)

type PageWrapper interface {
	page() contenthub.Page
}

// unwrapPage is used in equality checks and similar.
func unwrapPage(in any) (contenthub.Page, error) {
	switch v := in.(type) {
	case PageWrapper:
		return v.page(), nil
	case types.Unwrapper:
		return unwrapPage(v.Unwrapv())
	case contenthub.Page:
		return v, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unwrapPage: %T not supported", in)
	}
}

func MustUnwrapPage(in any) contenthub.Page {
	p, err := unwrapPage(in)
	if err != nil {
		panic(err)
	}

	return p
}
