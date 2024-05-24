package entity

import (
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

// PathSpec holds methods that decides how paths in URLs and files in Hugo should look like.
type PathSpec struct {
}

func (ps *PathSpec) LayoutFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Layouts.Fs
}

func (ps *PathSpec) ContentFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Content.Fs
}

func (ps *PathSpec) AssetsFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Assets.Fs
}

func (ps *PathSpec) AssetsFsRealFilename(rel string) string {
	return ps.BaseFs.SourceFilesystems.Assets.RealFilename(rel)
}

func (ps *PathSpec) AssetsFsRealDirs(from string) []string {
	return ps.BaseFs.SourceFilesystems.Assets.RealDirs(from)
}

func (ps *PathSpec) AssetsFsMakePathRelative(filename string, checkExists bool) (string, bool) {
	return ps.BaseFs.SourceFilesystems.Assets.MakePathRelative(filename, checkExists)
}

func (ps *PathSpec) ResourcesCacheFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.ResourcesCache
}

func (ps *PathSpec) WorkFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Work
}

// RelPathify trims any WorkingDir prefix from the given filename. If
// the filename is not considered to be absolute, the path is just cleaned.
func (ps *PathSpec) RelPathify(filename string, workingDir string) string {
	filename = filepath.Clean(filename)
	if !filepath.IsAbs(filename) {
		return filename
	}

	return strings.TrimPrefix(strings.TrimPrefix(filename, workingDir), paths.FilePathSeparator)
}
