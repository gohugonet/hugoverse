package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"sync"
)

type resourceAdapterInner struct {
	// The context that started this transformation.
	ctx context.Context

	target resources.TransformableResource

	stale.Staler

	//spec *Spec

	// Handles publishing (to /public) if needed.
	*publishOnce
}

type publishOnce struct {
	publisherInit sync.Once
	publisherErr  error
}
