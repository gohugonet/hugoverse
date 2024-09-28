package api

import (
	"log"
	"os"
	"path/filepath"
)

func adminStaticDir() string {
	staticDir := os.Getenv("HUGOVERSE_ADMIN_STATIC_DIR")
	if staticDir == "" {
		staticDir = filepath.Join(getWd(), "internal", "interfaces", "api", "admin", "static")
	}
	return staticDir
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}
