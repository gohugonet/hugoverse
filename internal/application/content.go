package application

import (
	"fmt"
	configFact "github.com/mdfriday/hugoverse/internal/domain/config/factory"
	"github.com/mdfriday/hugoverse/internal/domain/content/entity"
	"github.com/mdfriday/hugoverse/internal/domain/content/factory"
	"github.com/mdfriday/hugoverse/internal/domain/content/repository"
	contentHubFact "github.com/mdfriday/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/mdfriday/hugoverse/internal/domain/fs/factory"
	mdFact "github.com/mdfriday/hugoverse/internal/domain/markdown/factory"
	moduleFact "github.com/mdfriday/hugoverse/internal/domain/module/factory"
	rsFact "github.com/mdfriday/hugoverse/internal/domain/resources/factory"
	siteFact "github.com/mdfriday/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/mdfriday/hugoverse/internal/domain/template/factory"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/database"
	"time"
)

func NewContentServer(db repository.Repository) *entity.Content {
	return factory.NewContent(db, &dir{})
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
		Image:      resources.Image,
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

	db, err := database.New(DataDir())
	if err != nil {
		return err
	}

	ct := factory.NewContentWithServices(db, &siteServices{
		Config:     c,
		Fs:         sfs,
		ContentHub: ch,
	}, &dir{})

	db.RegisterContentBuckets(ct.AllContentTypeNames())
	defer db.Close()

	err = ct.LoadHugoProject()

	fmt.Printf("sorting...")
	time.Sleep(3 * time.Second)
	fmt.Println(" done.")

	return err
}
