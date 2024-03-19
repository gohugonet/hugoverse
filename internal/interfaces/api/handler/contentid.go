package handler

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/search"
)

type contentIdentifier struct {
	contentType string
	id          string
}

func (c *contentIdentifier) ContentType() string {
	return c.contentType
}

func (c *contentIdentifier) ID() string {
	return c.id
}

func IndexToIdentifier(index search.Index) content.Identifier {
	return &contentIdentifier{
		contentType: index.Namespace(),
		id:          index.ID(),
	}
}

func ConvertToIdentifiers(indices []search.Index) []content.Identifier {
	var ids []content.Identifier
	for _, i := range indices {
		ids = append(ids, IndexToIdentifier(i))
	}
	return ids
}
