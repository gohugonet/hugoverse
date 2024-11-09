package api

import (
	"embed"
	"log"
	"net/http"
	"os"
)

//go:embed admin/static/*
var staticFiles embed.FS

func adminStaticDir() http.FileSystem {
	staticDir := os.Getenv("HUGOVERSE_ADMIN_STATIC_DIR")
	if staticDir == "" {
		return http.FS(staticFiles)
	}
	return http.Dir(staticDir)
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}
