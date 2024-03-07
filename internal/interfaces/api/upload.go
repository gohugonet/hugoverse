package api

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// StoreFiles stores file uploads at paths like /YYYY/MM/filename.ext
func (s *Server) StoreFiles(req *http.Request) (map[string]string, error) {
	err := req.ParseMultipartForm(1024 * 1024 * 4) // maxMemory 4MB
	if err != nil {
		return nil, err
	}

	ts := req.FormValue("timestamp") // timestamp in milliseconds since unix epoch

	if ts == "" {
		ts = timestamp.Now()
	}

	// To use for FormValue name:urlPath
	urlPaths := make(map[string]string)

	if len(req.MultipartForm.File) == 0 {
		return urlPaths, nil
	}

	req.Form.Set("timestamp", ts)

	tm, err := timestamp.ConvertToTime(ts)
	if err != nil {
		return nil, err
	}

	urlPathPrefix := "api"
	uploadDirName := "uploads"
	uploadDir := filepath.Join(uploadDir(), fmt.Sprintf("%d", tm.Year()), fmt.Sprintf("%02d", tm.Month()))
	err = os.MkdirAll(uploadDir, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, err
	}

	// loop over all files and save them to disk
	for name, fds := range req.MultipartForm.File {
		filename, err := s.contentApp.NormalizeString(fds[0].Filename)
		if err != nil {
			return nil, err
		}

		src, err := fds[0].Open()
		if err != nil {
			err := fmt.Errorf("Couldn't open uploaded file: %s", err)
			return nil, err

		}
		defer src.Close()

		// check if file at path exists, if so, add timestamp to file
		absPath := filepath.Join(uploadDir, filename)

		if _, err := os.Stat(absPath); !os.IsNotExist(err) {
			filename = fmt.Sprintf("%d-%s", time.Now().Unix(), filename)
			absPath = filepath.Join(uploadDir, filename)
		}

		// save to disk (TODO: or check if S3 credentials exist, & save to cloud)
		dst, err := os.Create(absPath)
		if err != nil {
			err := fmt.Errorf("Failed to create destination file for upload: %s", err)
			return nil, err
		}

		// copy file from src to dst on disk
		var size int64
		if size, err = io.Copy(dst, src); err != nil {
			err := fmt.Errorf("Failed to copy uploaded file to destination: %s", err)
			return nil, err
		}

		// add name:urlPath to req.PostForm to be inserted into db
		urlPath := fmt.Sprintf("/%s/%s/%d/%02d/%s", urlPathPrefix, uploadDirName, tm.Year(), tm.Month(), filename)
		urlPaths[name] = urlPath

		// add upload information to db
		go s.storeFileInfo(size, filename, urlPath, fds)
	}

	return urlPaths, nil
}

func (s *Server) storeFileInfo(size int64, filename, urlPath string, fds []*multipart.FileHeader) {
	data := url.Values{
		"name":           []string{filename},
		"path":           []string{urlPath},
		"content_type":   []string{fds[0].Header.Get("Content-Type")},
		"content_length": []string{fmt.Sprintf("%d", size)},
	}

	if err := s.adminApp.NewUpload(data); err != nil {
		s.Log.Errorf("Error saving file upload record to database: %s", err)
	}
}
