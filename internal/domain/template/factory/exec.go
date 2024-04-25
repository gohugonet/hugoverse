package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/entity"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
)

func New(fs template.Fs, cfs template.CustomizedFunctions) (template.Template, error) {
	b := newBuilder().
		withFs(fs).
		withNamespace(newNamespace()).
		withLookup(newLookup()).
		withCfs(cfs).
		buildFunctions().
		buildParser().
		buildExecutor()

	return b.build()
}

func newLookup() *entity.Lookup {
	return &entity.Lookup{
		BaseOf: valueobject.NewBaseOf(),
	}
}

func newNamespace() *entity.Namespace {
	return &entity.Namespace{
		StateMap: &valueobject.StateMap{
			Templates: make(map[string]*valueobject.State),
		},
	}
}
