package application

import (
	configAgr "github.com/gohugonet/hugoverse/internal/domain/config/entity"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	contentHubFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsAgr "github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	rsAgr "github.com/gohugonet/hugoverse/internal/domain/resources/entity"
	rsFact "github.com/gohugonet/hugoverse/internal/domain/resources/factory"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	siteAgr "github.com/gohugonet/hugoverse/internal/domain/site/entity"
	siteFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/gohugonet/hugoverse/internal/domain/template/factory"
)

func GenerateStaticSite() error {
	c, err := configFact.LoadConfig()
	if err != nil {
		return err
	}

	mods, err := moduleFact.New(c)
	if err != nil {
		return err
	}

	fs, err := fsFact.New(c, mods)
	if err != nil {
		return err
	}

	ch, err := contentHubFact.New(fs)
	if err != nil {
		return err
	}

	s := siteFact.New(fs, ch, &siteConfig{
		baseUrl:   c.BaseUrl(),
		languages: c.Languages(),
	})

	ws := &resourcesWorkspaceProvider{
		Config: c,
		Fs:     fs,
		Site:   s,
	}
	resources, err := rsFact.NewResources(ws)
	if err != nil {
		return err
	}

	exec, err := tmplFact.New(fs, &templateCustomizedFunctionsProvider{
		Markdown:   mdFact.NewMarkdown(),
		ContentHub: ch,
		Site:       s,
		Resources:  resources,
		Config:     c,
		Fs:         fs,
	})

	resources.SetupTemplateClient(exec) // Expose template service to resources operations

	if err != nil {
		return err
	}

	if err := ch.CollectPages(exec); err != nil {
		return err
	}

	if err := s.Build(exec); err != nil {
		return err
	}

	return nil
}

type resourcesWorkspaceProvider struct {
	*configAgr.Config
	*fsAgr.Fs
	*siteAgr.Site
}

type templateCustomizedFunctionsProvider struct {
	markdown.Markdown
	contenthub.ContentHub
	site.Site
	*rsAgr.Resources
	*configAgr.Config
	*fsAgr.Fs
}

type siteConfig struct {
	baseUrl   string
	languages []valueobject.LanguageConfig
}

func (s *siteConfig) BaseUrl() string {
	return s.baseUrl
}

func (s *siteConfig) Languages() []site.LanguageConfig {
	var langs []site.LanguageConfig
	for _, l := range s.languages {
		langs = append(langs, site.LanguageConfig(l))
	}
	return langs
}
