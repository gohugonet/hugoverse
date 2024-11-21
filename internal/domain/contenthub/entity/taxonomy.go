package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

type Taxonomy struct {
	Views []contenthub.Taxonomy

	FsSvc contenthub.FsService
	Cache *Cache
}

func (t *Taxonomy) Assemble(pages *doctree.NodeShiftTree[*PageTreesNode], pb *PageBuilder) error {
	for _, viewName := range t.Views {
		key := t.PluralTreeKey(viewName.Plural())
		if v := pages.Get(key); v == nil {

			fmi := t.FsSvc.NewFileMetaInfo(key + "/_index.md")
			f := valueobject.NewFileInfo(fmi)

			ps, err := newPageSource(f, t.Cache)
			if err != nil {
				return err
			}

			p, err := pb.WithSource(ps).KindBuild()
			if err != nil {
				return err
			}

			pages.InsertIntoValuesDimension(key, newPageTreesNode(p))
		}
	}

	return nil
}

func (t *Taxonomy) IsTaxonomyPath(p string) bool {
	ta := t.getTaxonomy(p)

	if ta == nil {
		return false
	}

	return p == path.Join(t.PluralTreeKey(ta.Plural()), "_index.md")
}

func (t *Taxonomy) PluralTreeKey(plural string) string {
	return cleanTreeKey(plural)
}

func (t *Taxonomy) getTaxonomy(s string) (v contenthub.Taxonomy) {
	for _, n := range t.Views {
		if strings.HasPrefix(s, t.PluralTreeKey(n.Plural())) {
			return n
		}
	}
	return
}

func (t *Taxonomy) IsZero(v contenthub.Taxonomy) bool {
	if v == nil {
		return true
	}

	return v.Singular() == ""
}

// The home page is represented with the zero string.
// All other keys starts with a leading slash. No trailing slash.
// Slashes are Unix-style.
func cleanTreeKey(elem ...string) string {
	var s string
	if len(elem) > 0 {
		s = elem[0]
		if len(elem) > 1 {
			s = path.Join(elem...)
		}
	}
	s = strings.TrimFunc(s, trimCutsetDotSlashSpace)
	s = filepath.ToSlash(strings.ToLower(paths.Sanitize(s)))
	if s == "" || s == "/" {
		return ""
	}
	if s[0] != '/' {
		s = "/" + s
	}
	return s
}

var trimCutsetDotSlashSpace = func(r rune) bool {
	return r == '.' || r == '/' || unicode.IsSpace(r)
}
