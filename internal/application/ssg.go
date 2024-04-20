package application

import (
	"fmt"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
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

	fmt.Println(fs)
	return nil

	//ch, err := chFact.New(fs)
	//if err != nil {
	//	return err
	//}
	//
	//if err := ch.CollectPages(); err != nil {
	//	return err
	//}
	//
	//site := stFact.New(fs, ch)
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
