package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// StoreFiles stores file uploads at paths like /YYYY/MM/filename.ext
func (s *Handler) StoreFiles(req *http.Request) (map[string]string, error) {
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
	uploadDir := filepath.Join(s.uploadDir, s.db.UserDir(), fmt.Sprintf("%d", tm.Year()), fmt.Sprintf("%02d", tm.Month()))
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
			err := fmt.Errorf("couldn't open uploaded file: %s", err)
			s.log.Errorf("Error opening uploaded file: %s", err)
			return nil, err

		}

		// Overwrite file if it exists by default
		// support later : check if file at path exists, if so, add timestamp to file
		absPath := filepath.Join(uploadDir, filename)

		//if _, err := os.Stat(absPath); os.IsExist(err) {
		//	filename = fmt.Sprintf("%d-%s", time.Now().Unix(), filename)
		//	absPath = filepath.Join(uploadDir, filename)
		//}

		// save to disk
		// (TODO: or check if S3 credentials exist, & save to cloud)
		dst, err := os.Create(absPath)
		if err != nil {
			err := fmt.Errorf("failed to create destination file for upload: %s", err)
			s.log.Errorf("Error creating destination file for upload: %s", err)
			return nil, err
		}

		// copy file from src to dst on disk
		var size int64
		if size, err = io.Copy(dst, src); err != nil {
			_ = src.Close()
			_ = dst.Close()
			err := fmt.Errorf("failed to copy uploaded file to destination: %s", err)
			s.log.Errorf("Error copying uploaded file to destination: %s", err)
			return nil, err
		}

		// Close the source and destination files explicitly
		_ = src.Close()
		_ = dst.Close()

		// add name:urlPath to req.PostForm to be inserted into db
		urlPath := fmt.Sprintf("/%s/%s/%s/%d/%02d/%s", urlPathPrefix, uploadDirName, s.db.UserDir(), tm.Year(), tm.Month(), filename)
		urlPaths[name] = urlPath

		// add upload information to db
		go func() {
			s.storeFileInfo(size, filename, urlPath, fds)
		}()
	}

	return urlPaths, nil
}

func (s *Handler) storeFileInfo(size int64, filename, urlPath string, fds []*multipart.FileHeader) {
	data := url.Values{
		"name":           []string{filename},
		"path":           []string{urlPath},
		"content_type":   []string{fds[0].Header.Get("Content-Type")},
		"content_length": []string{fmt.Sprintf("%d", size)},
	}

	s.log.Debugln("storeFileInfo: ", filename, urlPath, fmt.Sprintf("%d", size))

	if err := s.adminApp.NewUpload(data); err != nil {
		s.log.Errorf("Error saving file upload record to database: %v", err)
	}
}

func (s *Handler) deleteUploadFromDisk(id string) error {
	// get data on file
	data, err := s.adminApp.GetUpload(id)
	if err != nil {
		return err
	}

	// unmarshal data
	upload := s.adminApp.UploadCreator()()
	if err = json.Unmarshal(data, &upload); err != nil {
		return err
	}

	ut, ok := upload.(admin.Traceable)
	if !ok {
		return fmt.Errorf("invalid upload type")
	}

	// split and rebuild path in OS friendly way
	// use path to delete the physical file from disk
	pathSplit := strings.Split(strings.TrimPrefix(ut.FilePath(), "/api/"), "/")
	pathJoin := filepath.Join(pathSplit...)
	err = os.Remove(pathJoin)
	if err != nil {
		return err
	}

	return nil
}
