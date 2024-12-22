package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"path/filepath"
	"strings"
)

func (s *Site) Params() maps.Params {
	return s.ConfigSvc.ConfigParams()
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
	if err != nil {
		return nil, err
	}

	return s.sitePage(p)
}

func (s *Site) Pages() []*Page {
	cps := s.ContentSvc.GlobalPages()

	var pages []*Page

	for _, cp := range cps {
		p, err := s.sitePage(cp)
		if err != nil {
			continue
		}
		pages = append(pages, p)
	}
	return pages
}

func (s *Site) RegularPages() []*Page {
	cps := s.ContentSvc.GlobalRegularPages()

	var pages []*Page

	for _, cp := range cps {
		p, err := s.sitePage(cp)
		if err != nil {
			continue
		}
		pages = append(pages, p)
	}
	return pages

}

func (s *Site) pageOutput(p contenthub.Page) (contenthub.PageOutput, error) {
	pos, err := p.PageOutputs()
	if err != nil {
		return nil, err
	}
	if len(pos) != 1 {
		return nil, fmt.Errorf("expected 1 page output, got %d", len(pos))
	}
	po := pos[0] // TODO, check for multiple outputs

	return po, nil
}

func (s *Site) sitePage(p contenthub.Page) (*Page, error) {
	po, err := s.pageOutput(p)
	if err != nil {
		return nil, err
	}

	sp := &Page{
		resSvc:    s.ResourcesSvc,
		tmplSvc:   s.Template,
		langSvc:   s.LanguageSvc,
		publisher: s.Publisher,
		git:       s.GitSvc,

		Page:       p,
		PageOutput: po,
		Site:       s,
	}

	sources, err := s.ContentSvc.GetPageSources(sp.Page)
	if err != nil {
		return nil, err
	}

	if err := sp.processResources(sources); err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *Site) siteWeightedPage(p contenthub.OrdinalWeightPage) (*WeightedPage, error) {
	sp, err := s.sitePage(p)
	if err != nil {
		return nil, err
	}

	return &WeightedPage{sp, p}, nil
}
