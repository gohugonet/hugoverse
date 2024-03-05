package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
)

type ContentServer struct {
	content.Content
}

func NewContentServer() *ContentServer {
	return &ContentServer{
		Content: factory.NewContent(),
	}
}
