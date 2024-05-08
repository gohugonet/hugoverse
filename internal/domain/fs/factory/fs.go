package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

var log = loggers.NewDefault()

func New(dir fs.Dir, mods module.Modules) (*entity.Fs, error) {
	f := &entity.Fs{
		OriginFs: NewOriginFs(dir),
	}

	bfs, err := NewBaseFS(dir, f.OriginFs, mods)
	if err != nil {
		return nil, err
	}

	f.PathSpec = &entity.PathSpec{
		BaseFs: bfs,
	}

	return f, nil
}
