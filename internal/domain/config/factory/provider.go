package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/entity"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"os"
	"path"
	"path/filepath"
)

const (
	DefaultThemesDir  = "themes"
	DefaultPublishDir = "public"
)

func LoadConfig() (*entity.Config, error) {
	currentDir, _ := os.Getwd()
	workingDir := filepath.Clean(currentDir)

	l := &ConfigLoader{
		SourceDescriptor: &sourceDescriptor{
			fs:       &afero.OsFs{},
			filename: path.Join(workingDir, "config.toml"),
		},
		Cfg: valueobject.NewDefaultProvider(),
		BaseDirs: valueobject.BaseDirs{
			WorkingDir: workingDir,
			ThemesDir:  paths.AbsPathify(workingDir, DefaultThemesDir),
			PublishDir: paths.AbsPathify(workingDir, DefaultPublishDir),
			CacheDir:   "",
		},
		Logger: loggers.NewDefault(),
	}

	defer l.deleteMergeStrategies()
	p, err := l.loadConfigByDefault()
	if err != nil {
		return nil, err
	}

	c := &entity.Config{
		ConfigSourceFs: l.SourceDescriptor.Fs(),
		Provider:       p,

		Root:      entity.Root{},
		Caches:    entity.Caches{},
		Security:  entity.Security{},
		Module:    entity.Module{},
		Language:  entity.Language{},
		Imaging:   entity.Imaging{},
		MediaType: entity.MediaType{},
	}

	if err := l.assembleConfig(c); err != nil {
		return nil, err
	}

	return c, nil
}
