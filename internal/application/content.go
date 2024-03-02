package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
	"log"
	"os"
)

type ContentServer struct {
	content content.Content
}

func NewContentServer() *ContentServer {
	return &ContentServer{
		content: factory.NewContent(),
	}
}

func (s *ContentServer) AllContentTypeNames() []string {
	return s.content.AllContentTypeNames()
}

func (s *ContentServer) DataDir() string {
	return dataDir()
}

func (s *ContentServer) GetContent(name string) (func() interface{}, bool) {
	return s.content.GetContent(name)
}

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
