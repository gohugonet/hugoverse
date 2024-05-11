package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/gohugonet/hugoverse/pkg/overlayfs"
	"github.com/spf13/afero"
)

type FilesystemsCollector struct {
	SourceProject afero.Fs // Source for project folders

	OverlayMounts        *overlayfs.OverlayFs
	OverlayMountsContent *overlayfs.OverlayFs
	OverlayResources     *overlayfs.OverlayFs

	OverlayFull *overlayfs.OverlayFs

	// Maps component type (layouts, static, content etc.) an ordered list of
	// directories representing the overlay filesystems above.
	OverlayDirs map[string][]FileMetaInfo
}

func (c *FilesystemsCollector) AddDirs(rfs *RootMappingFs) {
	for _, componentFolder := range module.ComponentFolders {
		c.addDir(rfs, componentFolder)
	}
}

func (c *FilesystemsCollector) addDir(rfs *RootMappingFs, componentFolder string) {
	dirs, err := rfs.Dirs(componentFolder)

	if err == nil { // event dirs is nil
		// merge all the same component folder from different rfs in the same array
		c.OverlayDirs[componentFolder] = append(c.OverlayDirs[componentFolder], dirs...)
	}
}
