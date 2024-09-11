package static

import (
	"net/http"
	"os"
)

type filesOnlyFs struct {
	fs http.FileSystem
}

func (fs filesOnlyFs) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return noDirFile{f}, nil
}

type noDirFile struct {
	http.File
}

func (f noDirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
