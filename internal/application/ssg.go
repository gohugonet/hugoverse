package application

import (
	cfFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	chFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
)

func GenerateStaticSite(projPath string) error {
	c, err := cfFact.NewConfigFromPath(projPath)
	if err != nil {
		return err
	}

	mods, err := mdFact.New(c.GetString("theme"))
	if err != nil {
		return err
	}

	fs, err := fsFact.New(&fsDir{
		workingDir: c.GetString("workingDir"),
		publishDir: c.GetString("publishDir"),
	}, mods)

	ch, err := chFact.New(fs)
	if err != nil {
		return err
	}

	if err := ch.Process(); err != nil {
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
