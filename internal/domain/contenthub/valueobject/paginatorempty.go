package valueobject

import "github.com/mdfriday/hugoverse/internal/domain/contenthub"

type PaginatorEmpty func() error

func (f PaginatorEmpty) Paginate(groups contenthub.PageGroups) (contenthub.Pager, error) {
	return nil, f()
}

func (f PaginatorEmpty) Paginator() (contenthub.Pager, error) {
	return nil, f()
}

func (f PaginatorEmpty) Current() contenthub.Pager { return nil }

func (f PaginatorEmpty) SetCurrent(current contenthub.Pager) {}
