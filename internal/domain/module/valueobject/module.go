package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/module"
	"github.com/spf13/afero"
	"path/filepath"
)

type Modules []*Module

type Module struct {
	AbsDir string
	Fs     afero.Fs

	Path   string
	Parent *Module

	MountDirs []Mount

	// Go Module supported only
	*GoModule
}

func (m *Module) ApplyMounts(moduleImport Import) error {
	mounts := moduleImport.Mounts

	if len(mounts) == 0 {
		for _, componentFolder := range module.ComponentFolders {
			sourceDir := filepath.Join(m.AbsDir, componentFolder)
			_, err := m.Fs.Stat(sourceDir)
			if err == nil {
				mounts = append(mounts, Mount{
					SourcePath: componentFolder,
					TargetPath: componentFolder,
				})
			}
		}
	}

	m.MountDirs = mounts
	return nil
}

func (m *Module) Owner() module.Module {
	return m.Parent
}

func (m *Module) Mounts() []module.Mount {
	var mounts []module.Mount
	for _, mount := range m.MountDirs {
		mounts = append(mounts, mount)
	}
	return mounts
}

func (m *Module) AppendMount(mount Mount) {
	m.MountDirs = append(m.MountDirs, mount)
}

func (m *Module) Dir() string {
	return m.AbsDir
}
