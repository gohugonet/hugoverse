package factory

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/entity"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

// NewOriginFs creates a new Fs.
func NewOriginFs(dir fs.Dir) *entity.OriginFs {
	if dir.PublishDir() == "" {
		panic("publishDir is empty")
	}

	workingDir := dir.WorkingDir()
	if workingDir == "." {
		workingDir = ""
	}

	// Sanity check
	if len(workingDir) < 2 {
		panic("workingDir is too short")
	}

	absPublishDir := paths.AbsPathify(workingDir, dir.PublishDir())
	if !strings.HasSuffix(absPublishDir, paths.FilePathSeparator) {
		absPublishDir += paths.FilePathSeparator
	}
	// If root, remove the second '/'
	if absPublishDir == "//" {
		absPublishDir = paths.FilePathSeparator
	}

	absResourcesDir := paths.AbsPathify(workingDir, dir.ResourceDir())
	if !strings.HasSuffix(absResourcesDir, paths.FilePathSeparator) {
		absResourcesDir += paths.FilePathSeparator
	}
	if absResourcesDir == "//" {
		absResourcesDir = paths.FilePathSeparator
	}

	osFs := afero.NewOsFs()

	return &entity.OriginFs{
		Source:             valueobject.NewBaseFs(osFs),
		PublishDir:         valueobject.NewBaseFs(getPublishFs(osFs, absPublishDir, dir.MediaTypes())),
		WorkingDirReadOnly: getWorkingDirFsReadOnly(osFs, workingDir),
		WorkingDirWritable: getWorkingDirFsWritable(osFs, workingDir),

		AbsWorkingDir:   workingDir,
		AbsPublishDir:   absPublishDir,
		AbsResourcesDir: absResourcesDir,
	}
}

func getPublishFs(base afero.Fs, publishDir string, mediaTypes media.Types) afero.Fs {
	pub := afero.NewBasePathFs(base, publishDir)

	hashBytesReceiverFunc := func(name string, match bool) {
		if !match {
			return
		}
		fmt.Println("publish fs hashBytesReceiverFunc: ", name)
	}

	hashBytesSHouldCheck := func(name string) bool {
		ext := strings.TrimPrefix(filepath.Ext(name), ".")
		return mediaTypes.IsTextSuffix(ext)
	}

	return valueobject.NewHasBytesReceiver(pub,
		hashBytesSHouldCheck, hashBytesReceiverFunc, []byte(resources.PostProcessPrefix))
}

func getWorkingDirFsReadOnly(base afero.Fs, workingDir string) afero.Fs {
	if workingDir == "" {
		return afero.NewReadOnlyFs(base)
	}
	return afero.NewBasePathFs(afero.NewReadOnlyFs(base), workingDir)
}

func getWorkingDirFsWritable(base afero.Fs, workingDir string) afero.Fs {
	if workingDir == "" {
		return base
	}
	return afero.NewBasePathFs(base, workingDir)
}
