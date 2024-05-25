package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"golang.org/x/text/unicode/norm"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func isWrite(flag int) bool {
	return flag&os.O_RDWR != 0 || flag&os.O_WRONLY != 0
}

func cleanName(name string) string {
	name = strings.Trim(filepath.Clean(name), paths.FilePathSeparator)
	if name == "." {
		name = ""
	}
	return name
}

func mapKey(name string) string {
	return paths.FilePathSeparator + name
}

func normalizeFilename(filename string) string {
	if filename == "" {
		return ""
	}
	if runtime.GOOS == "darwin" {
		// When a file system is HFS+, its filepath is in NFD form.
		return norm.NFC.String(filename)
	}
	return filename
}

func RealFilename(cfs *ComponentFs, rel string) string {
	fi, err := cfs.Stat(rel)
	if err != nil {
		return rel
	}
	if realfi, ok := fi.(fs.FileMetaInfo); ok {
		return realfi.FileName()
	}

	return rel
}

// RealDirs gets a list of absolute paths to directories starting from the given
// path.
func RealDirs(from string) []string {
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
