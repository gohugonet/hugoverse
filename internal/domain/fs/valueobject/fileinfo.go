package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
)

func NewFileInfo(fi os.FileInfo, filename string) *FileInfo {
	info := &FileInfo{
		FileInfo:           fi,
		normalizedFilename: normalizeFilename(filename),
	}

	if fi.IsDir() {
		fmt.Println("NewFileInfo JoinStatFunc to be done")
	}

	return info
}

type FileInfo struct {
	os.FileInfo

	normalizedFilename string
	OpenFunc           func() (afero.File, error)
}

func (fi *FileInfo) Name() string {
	return fi.FileInfo.Name()
}
