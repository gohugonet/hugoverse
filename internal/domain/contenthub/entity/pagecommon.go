package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/compare"
	"github.com/gohugonet/hugoverse/pkg/lazy"
	"sync"
)

type pageCommon struct {
	m *pageMeta

	// Lazily initialized dependencies.
	init *lazy.Init

	// All of these represents the common parts of a page.Page
	contenthub.FileProvider
	contenthub.PageMetaProvider

	compare.Eqer

	layoutDescriptor     valueobject.LayoutDescriptor
	layoutDescriptorInit sync.Once

	// The parsed page content.
	pageContent

	// Calculated an cached translation mapping key
	translationKey     string
	translationKeyInit sync.Once

	// Will only be set for bundled pages.
	parent *pageState
}
