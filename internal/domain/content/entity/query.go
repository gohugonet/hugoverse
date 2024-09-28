package entity

import (
	"errors"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/search"
)

func (c *Content) search(contentType string, query string) ([][]byte, error) {
	// execute search for query provided, if no index for type send 404
	indices, err := c.Search.TypeQuery(contentType, query, 10, 0)
	if errors.Is(err, search.ErrNoIndex) {
		c.Log.Errorf("Index for type %s not found", contentType)

		return nil, err
	}
	if err != nil {
		c.Log.Errorf("Error searching for type %s: %v", contentType, err)
		return nil, err
	}

	// respond with json formatted results
	bb, err := c.GetContents(indices)
	if err != nil {
		c.Log.Errorf("Error getting content: %v", err)
		return nil, err
	}

	return bb, nil
}
