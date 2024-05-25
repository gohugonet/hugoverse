package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

type Fs struct {
	*OriginFs

	Content    *valueobject.ComponentFs
	Data       *valueobject.ComponentFs
	I18n       *valueobject.ComponentFs
	Layouts    *valueobject.ComponentFs
	Archetypes *valueobject.ComponentFs
	Assets     *valueobject.ComponentFs

	AssetsWithDuplicatesPreserved *valueobject.ComponentFs

	// The work folder (may be a composite of project and theme components).
	Work afero.Fs

	// Writable filesystem on top the project's resources directory,
	// with any sub module's resource fs layered below.
	ResourcesCache afero.Fs

	RootFss []*valueobject.RootMappingFs
}

func (f *Fs) NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return valueobject.NewBasePathFs(source, path)
}

func (f *Fs) Glob(fs afero.Fs, pattern string, handle func(fi fs.FileMetaInfo) (bool, error)) error {
	return valueobject.Glob(fs, pattern, handle)
}

func (f *Fs) LayoutFs() afero.Fs {
	return f.Layouts
}

func (f *Fs) ContentFs() afero.Fs {
	return f.Content
}

func (f *Fs) AssetsFs() afero.Fs {
	return f.Assets
}

func (f *Fs) AssetsFsRealFilename(rel string) string {
	return valueobject.RealFilename(f.Assets, rel)
}

func (f *Fs) AssetsFsRealDirs(from string) []string {
	dirs := f.Assets.RealDirs(from)
	var filtered []string
	for _, dirname := range dirs {
		if _, err := f.OriginFs.Source.Stat(dirname); err == nil {
			filtered = append(filtered, dirname)
		}
	}

	return filtered
}

func (f *Fs) AssetsFsMakePathRelative(filename string, checkExists bool) (string, bool) {
	return ps.BaseFs.SourceFilesystems.Assets.MakePathRelative(filename, checkExists)
}

func (f *Fs) ResourcesCacheFs() afero.Fs {
	return f.ResourcesCache
}

func (f *Fs) WorkFs() afero.Fs {
	return f.Work
}

// RelPathify trims any WorkingDir prefix from the given filename. If
// the filename is not considered to be absolute, the path is just cleaned.
func (f *Fs) RelPathify(filename string, workingDir string) string {
	filename = filepath.Clean(filename)
	if !filepath.IsAbs(filename) {
		return filename
	}

	return strings.TrimPrefix(strings.TrimPrefix(filename, workingDir), paths.FilePathSeparator)
}
