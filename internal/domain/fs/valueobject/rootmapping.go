package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/radixtree"
	"path/filepath"
	"strings"
)

// RootMapping describes a virtual file or directory mount.
type RootMapping struct {
	// The virtual mount.
	From     string
	FromBase string // The base directory of the virtual mount. //TODO
	// The source directory or file.
	To string
	// The base of To. May be empty if an
	// absolute path was provided.
	ToBasedir string
	// Whether this is a mount in the main project.
	IsProject bool
	// The virtual mount point, e.g. "blog".
	path string

	Meta *FileMeta // File metadata (lang etc.)
	Fi   FileMetaInfo
}

func GetRms(t *radixtree.Tree, key string) []RootMapping {
	var mappings []RootMapping
	v, found := t.Get(key)
	if found {
		mappings = v.([]RootMapping)
	}
	return mappings
}

func (r RootMapping) Filename(name string) string { // convert to disk absolute path
	if name == "" {
		return r.To
	}
	return filepath.Join(r.To, strings.TrimPrefix(name, r.From))
}
