package application

import (
	"fmt"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
	contentHubFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	rsFact "github.com/gohugonet/hugoverse/internal/domain/resources/factory"
	siteFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/gohugonet/hugoverse/internal/domain/template/factory"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/database"
	"time"
)

func NewContentServer(db repository.Repository) *entity.Content {
	return factory.NewContent(db, SearchDir())
}

func LoadHugoProject() error {
	c, err := configFact.LoadConfig()
	if err != nil {
		return err
	}

	mods, err := moduleFact.New(c)
	if err != nil {
		return err
	}

	sfs, err := fsFact.New(c, mods)
	if err != nil {
		return err
	}

	ch, err := contentHubFact.New(&chServices{
		Config: c,
		Fs:     sfs,
		Module: mods,
	})
	if err != nil {
		return err
	}

	ws := &resourcesWorkspaceProvider{
		Config: c,
		Fs:     sfs,
	}
	resources, err := rsFact.NewResources(ws)
	if err != nil {
		return err
	}

	s := siteFact.New(&siteServices{
		Config:     c,
		Fs:         sfs,
		ContentHub: ch,
		Resources:  resources,
	})

	exec, err := tmplFact.New(sfs, &templateCustomizedFunctionsProvider{
		Markdown:   mdFact.NewMarkdown(),
		ContentHub: ch,
		Site:       s,
		Resources:  resources,
		Config:     c,
		Fs:         sfs,
	})

	resources.SetupTemplateClient(exec) // Expose template service to resources operations

	if err != nil {
		return err
	}

	if err := ch.ProcessPages(exec); err != nil {
		return err
	}

	db := database.New(DataDir())
	ct := factory.NewContentWithServices(db, SearchDir(), &siteServices{
		Config:     c,
		Fs:         sfs,
		ContentHub: ch,
	})

	db.Start(ct.AllContentTypeNames())
	defer db.Close()

	err = ct.LoadHugoProject()

	fmt.Printf("sorting...")
	time.Sleep(3 * time.Second)
	fmt.Println(" done.")

	return err
}
