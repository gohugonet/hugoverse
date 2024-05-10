package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"io"
)

type hashProvider interface {
	Hash() string
}

type ResourceTransformable interface {
	resources.Resource
	Transformer
}

type Transformer interface {
	Transform(...ResourceTransformation) (ResourceTransformable, error)
	TransformWithContext(context.Context, ...ResourceTransformation) (ResourceTransformable, error)
}

// ResourceTransformation is the interface that a resource transformation step
// needs to implement.
type ResourceTransformation interface {
	Key() valueobject.ResourceTransformationKey
	Transform(ctx *valueobject.ResourceTransformationCtx) error
}

type Template interface {
	Parse(name, tpl string) (template.Preparer, error)
	ExecuteWithContext(ctx context.Context, templ template.Preparer, wr io.Writer, data any) error
}
