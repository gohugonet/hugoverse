package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
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
