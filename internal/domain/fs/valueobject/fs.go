package valueobject

import (
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
