package factory

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/module"
)

func NewBaseFS(dir fs.Dir, ofs fs.OriginFs, mods module.Modules) (*valueobject.BaseFs, error) {
	//publishFs := NewBaseFileDecorator(ofs.Publish())
	//sourceFs := NewBaseFileDecorator(afero.NewBasePathFs(ofs.Origin(), dir.WorkingDir()))

	b := &valueobject.BaseFs{
		//SourceFs:  sourceFs,
		//WorkDir:   ofs.Working(),
		//PublishFs: publishFs,
	}

	builder := &sourceFilesystemsBuilder{
		modules:  mods,
		sourceFs: NewBaseFileDecorator(ofs.Origin()),
		theBigFs: b.TheBigFs,
		result:   &valueobject.SourceFilesystems{}}

	sourceFilesystems, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build filesystems: %w", err)
	}

	b.SourceFilesystems = sourceFilesystems
	b.TheBigFs = builder.theBigFs

	return b, nil
}
