package application

import (
	"errors"
	configAgr "github.com/gohugonet/hugoverse/internal/domain/config/entity"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	chAgr "github.com/gohugonet/hugoverse/internal/domain/contenthub/entity"
	contentHubFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsAgr "github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
	moduleAgr "github.com/gohugonet/hugoverse/internal/domain/module/entity"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	rsAgr "github.com/gohugonet/hugoverse/internal/domain/resources/entity"
	rsFact "github.com/gohugonet/hugoverse/internal/domain/resources/factory"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	siteAgr "github.com/gohugonet/hugoverse/internal/domain/site/entity"
	siteFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/gohugonet/hugoverse/internal/domain/template/factory"
	"github.com/spf13/afero"
	"os"
	"sort"
)

var publishDirFs afero.Fs

func ServeGenerateStaticSite() (afero.Fs, error) {
	if err := GenerateStaticSite(); err != nil {
		return nil, err
	}

	return publishDirFs, nil
}

func GenerateStaticSiteWithTarget(target string) error {
	info, err := os.Stat(target)

	if os.IsNotExist(err) {
		return errors.New("file not exist")
	}

	if !info.IsDir() {
		return errors.New("target is not a directory")
	}

	if err := os.Chdir(target); err != nil {
		return err
	}

	return GenerateStaticSite()
}

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

	publishDirFs = fs.PublishDirStatic()

	go func() {
		_ = staticCopy(fs.Static, fs.PublishDirStatic())
	}()

	ch, err := contentHubFact.New(&chServices{
		Config: c,
		Fs:     fs,
		Module: mods,
	})
	if err != nil {
		return err
	}

	ws := &resourcesWorkspaceProvider{
		Config: c,
		Fs:     fs,
	}
	resources, err := rsFact.NewResources(ws)
	if err != nil {
		return err
	}

	s := siteFact.New(&siteServices{
		Config:     c,
		Fs:         fs,
		ContentHub: ch,
		Resources:  resources,
	})

	exec, err := tmplFact.New(fs, &templateCustomizedFunctionsProvider{
		Markdown:   mdFact.NewMarkdown(),
		ContentHub: ch,
		Site:       s,
		Resources:  resources,
		Image:      resources.Image,
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
}

type templateCustomizedFunctionsProvider struct {
	markdown.Markdown
	*chAgr.ContentHub
	*siteAgr.Site
	*rsAgr.Resources
	*rsAgr.Image
	*configAgr.Config
	*fsAgr.Fs
}

type chServices struct {
	*configAgr.Config
	*fsAgr.Fs
	*moduleAgr.Module
}

func (s *chServices) Views() []contenthub.Taxonomy {
	var t []contenthub.Taxonomy
	for _, v := range s.Config.Views() {
		t = append(t, taxonomy{
			singular: v.Singular,
			plural:   v.Plural,
		})
	}
	sort.Slice(t, func(i, j int) bool {
		return t[i].Singular() < t[j].Singular()
	})
	return t
}

type taxonomy struct {
	singular string
	plural   string
}

func (t taxonomy) Singular() string {
	return t.singular
}

func (t taxonomy) Plural() string {
	return t.plural
}

type siteServices struct {
	*configAgr.Config
	*fsAgr.Fs
	*chAgr.ContentHub
	*rsAgr.Resources
}

func (s *siteServices) Menus() map[string][]site.Menu {
	siteMenus := make(map[string][]site.Menu)

	ms := s.Config.AllMenus()

	for k, v := range ms {
		if siteMenus[k] == nil {
			siteMenus[k] = make([]site.Menu, 0)
		}

		var menus []site.Menu
		for _, menu := range v {
			menus = append(menus, &siteMenu{
				name:   menu.Name,
				url:    menu.URL,
				weight: menu.Weight,
			})
		}

		siteMenus[k] = append(siteMenus[k], menus...)
	}

	return siteMenus
}

type siteMenu struct {
	name   string
	url    string
	weight int
}

func (s *siteMenu) Name() string {
	return s.name
}

func (s *siteMenu) URL() string {
	return s.url
}

func (s *siteMenu) Weight() int {
	return s.weight
}
