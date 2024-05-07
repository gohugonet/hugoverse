package entity

import (
	"context"
)

type ResourceTransformer struct {
	Resource

	*resourceTransformations
}

func (r *ResourceTransformer) Transform(t ...ResourceTransformation) (ResourceTransformable, error) {
	return r.TransformWithContext(context.Background(), t...)
}

func (r *ResourceTransformer) TransformWithContext(ctx context.Context, t ...ResourceTransformation) (ResourceTransformable, error) {
	r.resourceTransformations = &resourceTransformations{
		transformations: append(r.transformations, t...),
	}

	return r, nil
}
