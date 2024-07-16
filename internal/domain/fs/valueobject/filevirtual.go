package valueobject

import (
	"github.com/spf13/afero"
	"os"
	"sync"
)

var (
	vf *fileVirtual
)

// GetVirtualFileInfo returns the FileInfo of the virtual file
func GetVirtualFileInfo() (os.FileInfo, error) {
	if vf == nil {
		vf = &fileVirtual{
			mfs:      afero.NewMemMapFs(),
			fileName: "/virtual/file.txt",
		}
	}
	return vf.GetFileInfo()
}

type fileVirtual struct {
	mfs      afero.Fs
	fileName string
	fileOnce sync.Once
	fileInfo os.FileInfo
	fileErr  error
}

func (fv *fileVirtual) createFile() {
	file, err := fv.mfs.Create(fv.fileName)
	if err != nil {
		fv.fileErr = err
		return
	}

	_, err = file.WriteString("This is a virtual file.")
	if err != nil {
		fv.fileErr = err
		return
	}

	file.Close()

	fileInfo, err := fv.mfs.Stat(fv.fileName)
	if err != nil {
		fv.fileErr = err
		return
	}

	fv.fileInfo = fileInfo
}

func (fv *fileVirtual) GetFileInfo() (os.FileInfo, error) {
	fv.fileOnce.Do(fv.createFile)
	return fv.fileInfo, fv.fileErr
}
