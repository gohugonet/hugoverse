package valueobject

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

type pageInfo struct {
	name     string
	kind     string
	dir      string
	sections []string
	buffer   *bytes.Buffer
}

func NewPageInfo(name, kind, dir string, sections []string, buf *bytes.Buffer) contenthub.PageInfo {
	return &pageInfo{
		name:     name,
		kind:     kind,
		dir:      dir,
		sections: sections,
		buffer:   buf,
	}
}

func (pi *pageInfo) Kind() string          { return pi.kind }
func (pi *pageInfo) Sections() []string    { return pi.sections }
func (pi *pageInfo) Dir() string           { return pi.dir }
func (pi *pageInfo) Name() string          { return pi.name }
func (pi *pageInfo) Buffer() *bytes.Buffer { return pi.buffer }
