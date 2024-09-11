package application

import (
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
	siteAgr "github.com/gohugonet/hugoverse/internal/domain/site/entity"
	siteFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/gohugonet/hugoverse/internal/domain/template/factory"
	"github.com/gohugonet/hugoverse/pkg/testkit"
	"github.com/spf13/afero"
	"os"
	"sort"
)

var publishDirFs afero.Fs

func ServeGenerateStaticSite() (afero.Fs, error) {
	tmpDir, _, err := testkit.MkTestSite()
	//defer clean()

	if err != nil {
		return nil, err
	}

	if err := os.Chdir(tmpDir); err != nil {
		return nil, err
	}

	if err := GenerateStaticSite(); err != nil {
		return nil, err
	}

	return publishDirFs, nil
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

	staticSvc := newStatic(fs.Static, fs.PublishDirStatic())
	go staticSvc.copyStatic()

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
