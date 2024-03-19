package api

import (
	"log"
	"os"
	"path/filepath"
)

func tlsDir() string {
	tlsDir := os.Getenv("HUGOVERSE_TLS_DIR")
	if tlsDir == "" {
		tlsDir = filepath.Join(dataDir(), "tls")
	}
	return tlsDir
}

func dataDir() string {
	dataDir := os.Getenv("HUGOVERSE_DATA_DIR")
	if dataDir == "" {
		return getWd()
	}
	return dataDir
}

func adminStaticDir() string {
	staticDir := os.Getenv("HUGOVERSE_ADMIN_STATIC_DIR")
	if staticDir == "" {
		staticDir = filepath.Join(getWd(), "internal", "interfaces", "api", "admin", "static")
	}
	return staticDir
}

func uploadDir() string {
	uploadDir := os.Getenv("HUGOVERSE_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(dataDir(), "uploads")
	}
	return uploadDir
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}

func searchDir() string {
	searchDir := os.Getenv("HUGOVERSE_SEARCH_DIR")
	if searchDir == "" {
		searchDir = filepath.Join(dataDir(), "search")
	}
	return searchDir
}
