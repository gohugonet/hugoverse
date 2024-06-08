package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/paths"
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
	To     string
	ToBase string // The base of To. May be empty if an absolute path was provided.

	ToFi *FileInfo
}

func GetRms(t *radixtree.Tree, key string) []RootMapping {
	var mappings []RootMapping
	v, found := t.Get(key)
	if found {
		mappings = v.([]RootMapping)
	}
	return mappings
}

func (rm RootMapping) Filename(name string) string { // convert to disk absolute path
	if name == "" {
		return rm.To
	}
	return filepath.Join(rm.To, strings.TrimPrefix(name, rm.From))
}

func (rm RootMapping) clean() {
	rm.From = strings.Trim(filepath.Clean(rm.From), paths.FilePathSeparator)
	rm.To = filepath.Clean(rm.To)
}

func (rm RootMapping) absFilename(name string) string {
	if name == "" {
		return rm.To
	}
	return filepath.Join(rm.To, strings.TrimPrefix(name, rm.From))
}

func (rm RootMapping) path() string {
	return strings.TrimPrefix(strings.TrimPrefix(rm.From, rm.FromBase), paths.FilePathSeparator)
}
