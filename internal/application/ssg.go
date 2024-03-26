package application

import (
	"fmt"
	cfFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	chFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	"path"
)

func GenerateStaticSite(projPath string) error {
	c, err := cfFact.NewConfigFromPath(path.Join(projPath, "config.toml"))
	if err != nil {
		return err
	}

	ch, err := chFact.New(
		&themeProvider{
			name: c.GetString("theme"),
		},
	)
	if err != nil {
		return err
	}

	fs, err := fsFact.New(&fsDir{
		workingDir: c.GetString("workingDir"),
		publishDir: c.GetString("publishDir"),
	}, ch.Modules)

	if err != nil {
		return err
	}

	fmt.Println(ch, fs)

	return nil
}

type themeProvider struct {
	name string
}

func (t *themeProvider) Name() string {
	return t.name
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
