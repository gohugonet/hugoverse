package application

import (
	cfFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	chFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	stFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
)

func GenerateStaticSite(projPath string) error {
	c, err := cfFact.LoadConfig()
	if err != nil {
		return err
	}

	mods, err := mdFact.New(c.Theme())
	if err != nil {
		return err
	}

	fs, err := fsFact.New(&fsDir{
		workingDir: c.WorkingDir(),
		publishDir: c.PublishDir(),
	}, mods)

	ch, err := chFact.New(fs)
	if err != nil {
		return err
	}

	if err := ch.CollectPages(); err != nil {
		return err
	}

	site := stFact.New(fs, ch)
	if err := site.Build(); err != nil {
		return err
	}

	return nil
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
