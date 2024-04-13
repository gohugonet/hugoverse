package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/maps"
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

// NewDefaultProvider creates a Provider backed by an empty maps.Params.
func NewDefaultProvider() config.Provider {
	return &valueobject.DefaultConfigProvider{
		Root: make(maps.Params),
	}
}

func LoadConfig() (config.Config, error) {
	currentDir, _ := os.Getwd()
	workingDir := filepath.Clean(currentDir)

	l := &ConfigLoader{
		SourceDescriptor: &sourceDescriptor{
			fs:       &afero.OsFs{},
			filename: path.Join(workingDir, "config.toml"),
		},
		Cfg: NewDefaultProvider(),
		BaseDirs: valueobject.BaseDirs{
			WorkingDir: workingDir,
			ThemesDir:  paths.AbsPathify(workingDir, DefaultThemesDir),
			PublishDir: paths.AbsPathify(workingDir, DefaultPublishDir),
		},
		Logger: loggers.NewDefault(),
	}

	defer l.deleteMergeStrategies()
	if err := l.loadConfigMain(); err != nil {
		return nil, err
	}

	c, err := l.loadConfigAggregator()
	if err != nil {
		return nil, err
	}

	return c, nil
}
