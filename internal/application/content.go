package application

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
)

type ContentServer struct {
	content.Content
}

func NewContentServer(db repository.Repository) *ContentServer {
	return &ContentServer{
		Content: factory.NewContent(db),
	}
}
