package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path"
	"strings"
	"sync"
)

const slash = "/"

type Target struct {
	Prefix string

	// Where to store the file on disk relative to the publish dir. OS slashes.
	FilePath string
	// The directory to write sub-resources of the above.
	SubResourceBaseTarget string

	Format output.Format
}

func (t *Target) TargetFilePath() string {
	return t.FilePath
}

func (t *Target) TargetSubResourceDir() string {
	return t.SubResourceBaseTarget
}

func (t *Target) TargetPrefix() string {
	return t.Prefix
}

// PagePathBuilder When adding state here, remember to update putPagePathBuilder.
type PagePathBuilder struct {
	els []string

	format output.Format

	// Builder state.
	IsUgly             bool
	BaseNameSameAsType bool // Remove it, only sitemap has same basename from both output format and descriptor
	NoSubResources     bool
	FullSuffix         string // File suffix including any ".".
	prefixLink         string
	PrefixPath         string
	LinkUpperOffset    int
}

func (p *PagePathBuilder) Add(el ...string) {
	// Filter empty and slashes.
	n := 0
	for _, e := range el {
		if e != "" && e != slash {
			el[n] = e
			n++
		}
	}
	el = el[:n]

	p.els = append(p.els, el...)
}

func (p *PagePathBuilder) ConcatLast(s string) {
	if len(p.els) == 0 {
		p.Add(s)
		return
	}
	old := p.els[len(p.els)-1]
	if old == "" {
		p.els[len(p.els)-1] = s
		return
	}
	if old[len(old)-1] == '/' {
		old = old[:len(old)-1]
	}
	p.els[len(p.els)-1] = old + s
}

func (p *PagePathBuilder) IsHtmlIndex() bool {
	return p.Last() == "index.html"
}

func (p *PagePathBuilder) Last() string {
	if p.els == nil {
		return ""
	}
	return p.els[len(p.els)-1]
}

func (p *PagePathBuilder) Link() string {
	link := p.Path(p.LinkUpperOffset)

	// TODO: when we create link, we should look into this again
	//if p.BaseNameSameAsType {
	//	link = strings.TrimSuffix(link, p.d.BaseName)
	//}

	if p.prefixLink != "" {
		link = "/" + p.prefixLink + link
	}

	if p.LinkUpperOffset > 0 && !strings.HasSuffix(link, "/") {
		link += "/"
	}

	return link
}

func (p *PagePathBuilder) LinkDir() string {
	if p.NoSubResources {
		return ""
	}

	pathDir := p.PathDirBase()

	if p.prefixLink != "" {
		pathDir = "/" + p.prefixLink + pathDir
	}

	return pathDir
}

func (p *PagePathBuilder) Path(upperOffset int) string {
	upper := len(p.els)
	if upperOffset > 0 {
		upper -= upperOffset
	}
	pth := path.Join(p.els[:upper]...)
	return paths.AddLeadingSlash(pth)
}

func (p *PagePathBuilder) PathDir() string {
	dir := p.PathDirBase()
	if p.PrefixPath != "" {
		dir = "/" + p.PrefixPath + dir
	}
	return dir
}

func (p *PagePathBuilder) PathDirBase() string {
	if p.NoSubResources {
		return ""
	}

	dir := p.Path(0)
	isIndex := strings.HasPrefix(p.Last(), p.format.BaseName+".")

	if isIndex {
		dir = paths.Dir(dir)
	} else {
		dir = strings.TrimSuffix(dir, p.FullSuffix)
	}

	if dir == "/" {
		dir = ""
	}

	return dir
}

func (p *PagePathBuilder) PathFile() string {
	dir := p.Path(0)
	if p.PrefixPath != "" {
		dir = "/" + p.PrefixPath + dir
	}
	return dir
}

func (p *PagePathBuilder) Prepend(el ...string) {
	p.els = append(p.els[:0], append(el, p.els[0:]...)...)
}

func (p *PagePathBuilder) Sanitize() {
	for i, el := range p.els {
		p.els[i] = strings.ToLower(paths.Sanitize(el))
	}
}

var pagePathBuilderPool = &sync.Pool{
	New: func() any {
		return &PagePathBuilder{}
	},
}

func GetPagePathBuilder(format output.Format) *PagePathBuilder {
	b := pagePathBuilderPool.Get().(*PagePathBuilder)
	b.format = format
	return b
}

func PutPagePathBuilder(b *PagePathBuilder) {
	b.els = b.els[:0]
	b.FullSuffix = ""
	b.BaseNameSameAsType = false
	b.IsUgly = false
	b.NoSubResources = false
	b.prefixLink = ""
	b.PrefixPath = ""
	b.LinkUpperOffset = 0
	pagePathBuilderPool.Put(b)
}
