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

	// When in multihost we have one static filesystem per language. The sync
	// static files is currently done outside of the Hugo build (where there is
	// a concept of a site per language).
	// When in non-multihost mode there will be one entry in this map with a blank key.
	Static map[string]*valueobject.ComponentFs

	// Writable filesystem on top the project's resources directory,
	// with any sub module's resource fs layered below.
	ResourcesCache afero.Fs

	RootFss []*valueobject.RootMappingFs

	*Service
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

func (f *Fs) ReverseLookupContent(filename string, checkExists bool) ([]fs.ComponentPath, error) {
	cps, err := f.Content.ReverseLookup(filename, checkExists)
	if err != nil {
		return nil, err
	}

	var fcps []fs.ComponentPath
	for _, cp := range cps {
		fcps = append(fcps, cp)
	}

	return fcps, err
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
	return f.Assets.MakePathRelative(filename, checkExists)
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
