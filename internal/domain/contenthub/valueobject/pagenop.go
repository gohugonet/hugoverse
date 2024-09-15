package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

var (
	NilPage *nopPage
)

// PageNop implements Page, but does nothing.
type nopPage int

func (p *nopPage) Pages() contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) RawContent() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Section() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Title() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Description() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Params() maps.Params {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Kind() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsPage() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsSection() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PageIdentity() contenthub.PageIdentity {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PageFile() contenthub.File {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) MarkStale() {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsStale() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Paths() *paths.Path {
	//TODO implement me
	panic("implement me")
}
func (p *nopPage) Path() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Opener() pio.OpenReadSeekCloser {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Layouts() []string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PageOutputs() ([]contenthub.PageOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsAncestor(other contenthub.Page) bool {
	//TODO implement me
	panic("implement me")
}
