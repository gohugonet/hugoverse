package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Taxonomy struct {
	Taxonomies map[string]string

	views          []valueobject.ViewName
	viewsByTreeKey map[string]valueobject.ViewName
}

func (t Taxonomy) SetupViews() {
	var views []valueobject.ViewName
	for k, v := range t.Taxonomies {
		views = append(views, valueobject.ViewName{Singular: k, Plural: v, PluralTreeKey: cleanTreeKey(v)})
	}
	sort.Slice(views, func(i, j int) bool {
		return views[i].Plural < views[j].Plural
	})

	viewsByTreeKey := make(map[string]valueobject.ViewName)
	for _, v := range views {
		viewsByTreeKey[v.PluralTreeKey] = v
	}

	t.views = views
	t.viewsByTreeKey = viewsByTreeKey
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
