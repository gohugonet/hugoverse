package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
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

func (s *ContentServer) GetContent(name string) (func() interface{}, bool) {
	return s.content.GetContent(name)
}
