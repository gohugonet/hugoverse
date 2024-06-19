package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

type pageMeta struct {
	f        contenthub.File
	pathInfo *paths.Path

	bundled bool // Set if this page is bundled inside another.

	stale.Staler
}

func newBundledPageMeta(f contenthub.File, pathInfo *paths.Path) *pageMeta {
	m := &pageMeta{
		f:        f,
		pathInfo: pathInfo,

		bundled: true,
	}

	m.setupStaler()

	return m
}

func newPageMeta(f contenthub.File, pathInfo *paths.Path) *pageMeta {
	m := &pageMeta{
		f:        f,
		pathInfo: pathInfo,

		bundled: false,
	}
	return m
}

func (m *pageMeta) setupStaler() {
	m.Staler = &stale.AtomicStaler{}
}
