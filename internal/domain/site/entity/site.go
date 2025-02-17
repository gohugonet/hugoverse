package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/site"
	"github.com/mdfriday/hugoverse/internal/domain/site/valueobject"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"time"
)

type Site struct {
	ConfigSvc      site.ConfigService
	ContentSvc     site.ContentService
	TranslationSvc site.TranslationService
	ResourcesSvc   site.ResourceService
	LanguageSvc    site.LanguageService
	Sitemap        site.SitemapService

	GitSvc *valueobject.GitMap

	Publisher *Publisher

	Template site.Template

	*valueobject.Author
	*valueobject.Compiler

	Title string

	*URL
	*Ref
	*Language
	*Navigation
	*Reserve

	home *Page

	Log     loggers.Logger `json:"-"`
	siteLog logg.LevelLogger

	// Lazily loaded site dependencies
	lazy *siteInit
}

func (s *Site) Build(t site.Template) error {
	s.siteLog = s.Log.InfoCommand("site build")
	defer loggers.TimeTrackf(s.siteLog, time.Now(), nil, "")

	s.Template = t

	if err := s.setup(); err != nil {
		return err
	}
	for _, l := range s.LangSvc.LanguageKeys() {
		s.Language.currentLanguage = l
		err := s.render()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Site) setup() error {
	l := s.siteLog.WithField("step setup", "setup url and languages")
	start := time.Now()
	defer func() {
		loggers.TimeTrackf(l, start, nil, "")
	}()

	if err := s.Template.MarkReady(); err != nil {
		return err
	}

	if err := s.URL.setup(); err != nil {
		return err
	}
	if err := s.Language.setup(); err != nil {
		return err
	}
	return nil
}

func (s *Site) render() error {
	l := s.siteLog.WithField("step render", "render sties")
	start := time.Now()
	defer func() {
		loggers.TimeTrackf(l, start, nil, "")
	}()

	spc.clear()

	if err := s.renderPages(); err != nil {
		return err
	}

	return nil
}

func (s *Site) renderPages() error {
	render := newRender(s.siteLog)
	go render.startRenderPages()

	if err := s.ContentSvc.WalkPages(s.Language.CurrentLanguageIndex(), func(p contenthub.Page) error {
		sitePage := &Page{
			resSvc:    s.ResourcesSvc,
			tmplSvc:   s.Template,
			langSvc:   s.LanguageSvc,
			publisher: s.Publisher,
			git:       s.GitSvc,

			Page: p,
			Site: s,
		}

		po, err := s.pageOutput(p)
		if err != nil {
			return err
		}
		sitePage.PageOutput = po

		sources, err := s.ContentSvc.GetPageSources(sitePage.Page)
		if err != nil {
			return err
		}

		if err := sitePage.processResources(sources); err != nil {
			return err
		}

		if sitePage.Page.IsHome() {
			s.home = sitePage
		}

		render.pages <- sitePage

		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk pages: %w", herrors.ImproveIfNilPointer(err))
	}

	render.close()

	err := <-render.errs
	if err != nil {
		return fmt.Errorf("failed to render pages: %w", herrors.ImproveIfNilPointer(err))
	}

	return nil
}
