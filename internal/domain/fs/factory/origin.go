package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	"github.com/spf13/afero"
	"os"
)

// NewOriginFs creates a new Fs.
func NewOriginFs(dir fs.Dir) *entity.OriginFs {
	afs := afero.NewOsFs()
	workingFs := afero.NewBasePathFs(afs, dir.WorkingDir())

	// Make sure we always have the /public folder ready to use.
	if err := workingFs.MkdirAll(dir.PublishDir(), 0777); err != nil && !os.IsExist(err) {
		panic(err)
	}
	pubFs := afero.NewBasePathFs(workingFs, dir.PublishDir())

	return &entity.OriginFs{
		Source:             afs,
		PublishDir:         pubFs,
		WorkingDirReadOnly: getWorkingDirFsReadOnly(workingFs, dir.WorkingDir()),
	}
}

func getWorkingDirFsReadOnly(base afero.Fs, workingDir string) afero.Fs {
	if workingDir == "" {
		return afero.NewReadOnlyFs(base)
	}
	return afero.NewBasePathFs(afero.NewReadOnlyFs(base), workingDir)
}
