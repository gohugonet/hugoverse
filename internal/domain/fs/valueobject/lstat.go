package valueobject

import (
	"github.com/spf13/afero"
	"os"
)

// LstatIfPossible if the filesystem supports it, use Lstat, else use fs.Stat
func LstatIfPossible(fs afero.Fs, path string) (os.FileInfo, bool, error) {
	if lfs, ok := fs.(afero.Lstater); ok {
		fi, b, err := lfs.LstatIfPossible(path)
		return fi, b, err
	}
	fi, err := fs.Stat(path)
	return fi, false, err
}
