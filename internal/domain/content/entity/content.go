package entity

import (
	"encoding/json"
	"errors"
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
	"github.com/gohugonet/hugoverse/pkg/form"
	"github.com/gorilla/schema"
	"net/url"
)

type Content struct {
	Types map[string]func() interface{}
	Repo  repository.Repository
}

func (c *Content) AllContentTypeNames() []string {
	keys := make([]string, 0, len(c.Types))
	for k := range c.Types {
		keys = append(keys, k)
	}
	return keys
}

func (c *Content) GetContent(name string) (func() interface{}, bool) {
	t, ok := c.Types[name]
	return t, ok
}

func (c *Content) NewContent(contentType string, data url.Values) (string, error) {
	t, ok := c.GetContent(contentType)
	if !ok {
		return "", errors.New("invalid content type")
	}
	ci := t()

	d, err := form.Convert(data)
	if err != nil {
		return "", err
	}
	// Decode Content
	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err = dec.Decode(ci, d)
	if err != nil {
		return "", err
	}

	cii, ok := ci.(content.Identifiable)
	if ok {
		uid, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		cii.SetUniqueID(uid)

		id, err := c.Repo.NextContentId(contentType)
		cii.SetItemID(int(id))
	} else {
		return "", errors.New("content type does not implement Identifiable")
	}

	slug, err := Slug(cii)
	if err != nil {
		return "", err
	}

	slug, err = c.Repo.CheckSlugForDuplicate(slug)
	if err != nil {
		return "", err
	}

	ciSlug, ok := ci.(content.Sluggable)
	if ok {
		ciSlug.SetSlug(slug)
	} else {
		return "", errors.New("content type does not implement Sluggable")
	}

	cis, ok := ci.(content.Statusable)
	if !ok {
		return "", errors.New("content type does not implement Statusable")
	}

	b, err := c.Marshal(cii)
	if err != nil {
		return "", err
	}

	if err := c.Repo.PutContent(
		cii.String(), ciSlug.ItemSlug(), contentType, cis.ItemStatus(), b); err != nil {
		return "", err
	}

	return cii.String(), nil
}

func (c *Content) Marshal(content any) ([]byte, error) {
	j, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *Content) AllContentTypes() map[string]func() interface{} {
	return c.Types
}

func (c *Content) NormalizeString(s string) (string, error) {
	return stringToSlug(s)
}
