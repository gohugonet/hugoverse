package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

// We create a pageOutput for every output format combination, even if this
// particular page isn't configured to be rendered to that format.
type pageOutput struct {
	// These interface provides the functionality that is specific for this
	// output format.
	contenthub.PagePerOutputProviders
	contenthub.ContentProvider

	// May be nil.
	cp *pageContentOutput
}

func newPageOutput() *pageOutput {
	po := &pageOutput{
		ContentProvider: nil,
	}

	return po
}

func (p *pageOutput) initContentProvider(cp *pageContentOutput) {
	if cp == nil {
		return
	}
	p.ContentProvider = cp
	p.cp = cp
}
