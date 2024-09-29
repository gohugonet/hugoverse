package entity

import "github.com/gohugonet/hugoverse/internal/domain/content/valueobject"

func (c *Content) BuildTarget(contentType, id, status string) (string, error) {
	site, err := c.getContent(contentType, id)
	if err != nil {
		return "", err
	}

	return site.(*valueobject.Site).WorkingDir, nil
}
