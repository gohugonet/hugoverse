package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
)

type pageMeta struct {
	f contenthub.File

	stale.Staler
}
