package application

import (
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	contentHubFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	"github.com/gohugonet/hugoverse/internal/domain/site"
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

	fs, err := fsFact.New(&fsDir{
		workingDir: c.WorkingDir(),
		publishDir: c.PublishDir(),
	}, mods)
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

	exec, err := tmplFact.New(fs, &templateCustomizedFunctionsProvider{
		Markdown:   mdFact.NewMarkdown(),
		ContentHub: ch,
		Site:       s,
	})

	if err != nil {
		return err
	}

	ch.SetTemplateExecutor(exec)
	if err := ch.CollectPages(); err != nil {
		return err
	}

	return nil
	//
	//if err := site.Build(); err != nil {
	//	return err
	//}
	//
	//return nil
}

type fsDir struct {
	workingDir string
	publishDir string
}

func (fs *fsDir) WorkingDir() string {
	return fs.workingDir
}
func (fs *fsDir) PublishDir() string {
	return fs.publishDir
}

type templateCustomizedFunctionsProvider struct {
	markdown.Markdown
	contenthub.ContentHub
	site.Site
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
