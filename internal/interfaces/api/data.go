package api

import (
	"log"
	"os"
)

func dataDir() string {
	dataDir := os.Getenv("HUGOVERSE_DATA_DIR")
	if dataDir == "" {
		return getWd()
	}
	return dataDir
}

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory", err)
	}
	return wd
}
