package js

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/resource/resourcehelpers"
)

func New(client Client) (*Namespace, error) {
	return &Namespace{
		jsClient: client,
	}, nil
}

// Namespace provides template functions for the "resources" namespace.
type Namespace struct {
	jsClient Client
}

// Build processes the given Resource with ESBuild.
func (ns *Namespace) Build(args ...any) (resources.Resource, error) {
	var (
		r          resources.Resource
		m          map[string]any
		targetPath string
		err        error
		ok         bool
	)

	r, targetPath, ok = resourcehelpers.ResolveIfFirstArgIsString(args)

	if !ok {
		r, m, err = resourcehelpers.ResolveArgs(args)
		if err != nil {
			return nil, err
		}
	}

	if targetPath != "" {
		m = map[string]any{"targetPath": targetPath}
	}

	return ns.jsClient.ProcessJs(r, m)
}
