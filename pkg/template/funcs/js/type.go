package js

import "github.com/gohugonet/hugoverse/internal/domain/resources"

type Client interface {
	ProcessJs(res resources.Resource, opts map[string]any) (resources.Resource, error)
}
