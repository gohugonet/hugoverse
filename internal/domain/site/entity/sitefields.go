package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"path/filepath"
	"strings"
)

func (s *Site) Params() maps.Params {
	// TODO： duplicate params, remove in next major
	return maps.Params{}
}

func (s *Site) Home() *Page {
	return s.home
}

func (s *Site) IsMultiLingual() bool {
	return s.Language.isMultipleLanguage()
}

func (s *Site) GetPage(ref ...string) (*Page, error) {
	if len(ref) > 1 {
		// This was allowed in Hugo <= 0.44, but we cannot support this with the
		// new API. This should be the most unusual case.
		return nil, fmt.Errorf(`too many arguments to .Site.GetPage: %v. Use lookups on the form {{ .Site.GetPage "/posts/mypage-md" }}`, ref)
	}

	key := ref[0]
	key = filepath.ToSlash(key)
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}

	p, err := s.ContentSvc.GetPageFromPath(key)

	return &Page{
		resSvc:    s.ResourcesSvc,
		tmplSvc:   s.Template,
		publisher: s.Publisher,

		Page: p,
		Site: s,
	}, err
}

func (s *Site) GitInfo() valueobject.GitInfo {
	return valueobject.GitInfo{}
}
