package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"time"
)

var (
	NilPage *nopPage
)

// PageNop implements Page, but does nothing.
type nopPage int

func (p *nopPage) RegularPagesRecursive() contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PublishDate() time.Time {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) RelatedKeywords(cfg contenthub.IndexConfig) ([]contenthub.Keyword, error) {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Name() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Store() *maps.Scratch {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Scratch() *maps.Scratch {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PrevInSection() contenthub.Page {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) NextInSection() contenthub.Page {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Sections(langIndex int) contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsTranslated() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PageDate() time.Time {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Truncated() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Current() contenthub.Pager {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) SetCurrent(current contenthub.Pager) {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Paginator() (contenthub.Pager, error) {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Paginate(groups contenthub.PageGroups) (contenthub.Pager, error) {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) RegularPages() contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Parent() contenthub.Page {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Terms(langIndex int, taxonomy string) contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) ShouldList(global bool) bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) ShouldListAny() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) NoLink() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PageWeight() int {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) PureContent() string {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) IsHome() bool {
	//TODO implement me
	panic("implement me")
}

func (p *nopPage) Translations() contenthub.Pages {
	//TODO implement me
	panic("implement me")
}

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

func (p *nopPage) Eq(other contenthub.Page) bool {
	//TODO implement me
	panic("implement me")
}
