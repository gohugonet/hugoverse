package factory

import (
	"errors"
	"github.com/mdfriday/hugoverse/internal/domain/module"
	"github.com/mdfriday/hugoverse/internal/domain/module/entity"
	"github.com/mdfriday/hugoverse/internal/domain/module/valueobject"
	"github.com/mdfriday/hugoverse/pkg/env"
	"github.com/mdfriday/hugoverse/pkg/hexec"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"path/filepath"
)

func New(info module.LoadInfo) (*entity.Module, error) {
	if !checkGoModule(info.WorkingDir(), info.Fs()) {
		return nil, errors.New("go.mod file not found, go module hugo project supported only")
	}

	log := loggers.NewDefault()

	var envs []string
	env.SetEnvVars(&envs, "PWD", info.WorkingDir(), "GO111MODULE", "on")

	ms := &entity.Module{
		GoClient: &valueobject.GoClient{
			Exec:    hexec.New(),
			Dir:     info.WorkingDir(),
			Environ: envs,
			Logger:  log,
		},
		Fs:            info.Fs(),
		WorkingDir:    info.WorkingDir(),
		ModuleImports: info.ImportPaths(),
		PathService:   info,
		DirService:    info,
		Logger:        log,
	}

	if err := ms.Load(); err != nil {
		return nil, err
	}

	ms.Lang = entity.NewLang(ms.All())

	return ms, nil
}

func checkGoModule(workingDir string, fs afero.Fs) bool {
	n := filepath.Join(workingDir, valueobject.GoModFilename)
	goModEnabled, _ := afero.Exists(fs, n)
	return goModEnabled
}
