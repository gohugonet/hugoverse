package valueobject

import (
	"fmt"
	"github.com/spf13/afero"
	"math/rand"
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
			mfs:         afero.NewMemMapFs(),
			fileName:    "/content/file.txt",
			fileContent: "This is a virtual file.",
		}
	}
	return vf.GetFileInfo()
}

func GetVirtualFileInfoWithContent(content string) (*fileVirtual, error) {
	vf = &fileVirtual{
		mfs:         afero.NewMemMapFs(),
		fileName:    fmt.Sprintf("/content/file_%s.md", generateFileName()),
		fileContent: content,
	}

	return vf, nil
}

func generateFileName() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000)) // 生成六位数字
}

type fileVirtual struct {
	mfs         afero.Fs
	fileName    string
	fileOnce    sync.Once
	fileInfo    os.FileInfo
	fileErr     error
	fileContent string
}

func (fv *fileVirtual) Open() (afero.File, error) {
	return fv.mfs.Open(fv.fileName)
}

func (fv *fileVirtual) FullName() string {
	return fv.fileName
}

func (fv *fileVirtual) createFile() {
	file, err := fv.mfs.Create(fv.fileName)
	if err != nil {
		fv.fileErr = err
		return
	}

	_, err = file.WriteString(fv.fileContent)
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
