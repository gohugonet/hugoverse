package valueobject

import (
	"github.com/spf13/afero"
	"path/filepath"
)

// SourceFilesystems contains the different source file systems. These can be
// composite file systems (theme and project etc.), and they have all root
// set to the source type the provides: data, i18n, static, layouts.
type SourceFilesystems struct {
	Content    *SourceFilesystem // set
	Data       *SourceFilesystem
	I18n       *SourceFilesystem
	Layouts    *SourceFilesystem // set
	Archetypes *SourceFilesystem
	Assets     *SourceFilesystem

	// Writable filesystem on top the project's resources directory,
	// with any sub module's resource fs layered below.
	ResourcesCache afero.Fs

	// The work folder (may be a composite of project and theme components).
	Work afero.Fs

	// When in multihost we have one static filesystem per language. The sync
	// static files is currently done outside of the Hugo build (where there is
	// a concept of a site per language).
	// When in non-multihost mode there will be one entry in this map with a blank key.
	Static map[string]*SourceFilesystem

	// All the /static dirs (including themes/modules).
	StaticDirs []FileMetaInfo
}

// A SourceFilesystem holds the filesystem for a given source type in Hugo (data,
// i18n, layouts, static) and additional metadata to be able to use that filesystem
// in server mode.
type SourceFilesystem struct {
	// Name matches one in files.ComponentFolders
	Name string

	// This is a virtual composite filesystem. It expects path relative to a context.
	Fs afero.Fs

	// The source filesystem (usually the OS filesystem).
	SourceFs afero.Fs

	// This filesystem as separate root directories, starting from project and down
	// to the themes/modules.
	Dirs []FileMetaInfo

	// When syncing a source folder to the target (e.g. /public), this may
	// be set to publish into a subfolder. This is used for static syncing
	// in multihost mode.
	PublishFolder string
}

func (d *SourceFilesystem) RealFilename(rel string) string {
	fi, err := d.Fs.Stat(rel)
	if err != nil {
		return rel
	}
	if realfi, ok := fi.(FileMetaInfo); ok {
		return realfi.Meta().Filename
	}

	return rel
}

// RealDirs gets a list of absolute paths to directories starting from the given
// path.
func (d *SourceFilesystem) RealDirs(from string) []string {
	var dirnames []string
	for _, m := range d.mounts() {
		if !m.IsDir() {
			continue
		}
		dirname := filepath.Join(m.Meta().Filename, from)
		if _, err := d.SourceFs.Stat(dirname); err == nil {
			dirnames = append(dirnames, dirname)
		}
	}
	return dirnames
}

func (d *SourceFilesystem) mounts() []FileMetaInfo {
	var m []FileMetaInfo
	WalkFilesystems(d.Fs, func(fs afero.Fs) bool {
		if rfs, ok := fs.(*RootMappingFs); ok {
			mounts, err := rfs.Mounts(d.Name)
			if err == nil {
				m = append(m, mounts...)
			}
		}
		return false
	})

	return m
}

// MakePathRelative creates a relative path from the given filename.
func (d *SourceFilesystem) MakePathRelative(filename string, checkExists bool) (string, bool) {
	cps, err := d.ReverseLookup(filename, checkExists)
	if err != nil {
		panic(err)
	}
	if len(cps) == 0 {
		return "", false
	}

	return filepath.FromSlash(cps[0].Path), true
}

// ReverseLookup returns the component paths for the given filename.
func (d *SourceFilesystem) ReverseLookup(filename string, checkExists bool) ([]ComponentPath, error) {
	var cps []ComponentPath
	WalkFilesystems(d.Fs, func(fs afero.Fs) bool {
		if rfs, ok := fs.(ReverseLookupProvder); ok {
			if c, err := rfs.ReverseLookupComponent(d.Name, filename); err == nil {
				if checkExists {
					n := 0
					for _, cp := range c {
						if _, err := d.Fs.Stat(filepath.FromSlash(cp.Path)); err == nil {
							c[n] = cp
							n++
						}
					}
					c = c[:n]
				}
				cps = append(cps, c...)
			}
		}
		return false
	})
	return cps, nil
}

type ReverseLookupProvder interface {
	ReverseLookup(filename string) ([]ComponentPath, error)
	ReverseLookupComponent(component, filename string) ([]ComponentPath, error)
}
