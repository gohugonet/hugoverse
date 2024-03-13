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
	"strconv"
)

type Content struct {
	Types map[string]content.Creator
	Repo  repository.Repository
}

func (c *Content) AllContentTypeNames() []string {
	keys := make([]string, 0, len(c.Types))
	for k := range c.Types {
		keys = append(keys, k)
	}
	return keys
}

func (c *Content) GetContentCreator(name string) (content.Creator, bool) {
	t, ok := c.Types[name]
	return t, ok
}

func (c *Content) GetContent(contentType, id, status string) ([]byte, error) {
	return c.Repo.GetContent(GetNamespace(contentType, status), id)
}

func (c *Content) DeleteContent(contentType, id, status string) error {
	data, err := c.GetContent(contentType, id, status)
	if err != nil {
		return err
	}
	ct, ok := c.GetContentCreator(contentType)
	if !ok {
		return errors.New("invalid content type")
	}
	cti := ct()
	if err = json.Unmarshal(data, cti); err != nil {
		return err
	}

	return c.Repo.DeleteContent(GetNamespace(contentType, status), id, cti.(content.Sluggable).ItemSlug())
}

func (c *Content) UpdateContent(contentType string, data url.Values) error {
	t, ok := c.GetContentCreator(contentType)
	if !ok {
		return errors.New("invalid content type")
	}
	ci := t()

	d, err := form.Convert(data)
	if err != nil {
		return err
	}
	// Decode Content
	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err = dec.Decode(ci, d)
	if err != nil {
		return err
	}

	b, err := c.Marshal(ci)
	if err != nil {
		return err
	}

	if err := c.Repo.PutContent(ci, b); err != nil {
		return err
	}

	return nil
}

func (c *Content) NewContent(contentType string, data url.Values) (string, error) {
	t, ok := c.GetContentCreator(contentType)
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
		if err != nil {
			return "", err
		}
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
	if ok {
		if cis.ItemStatus() == "" {
			cis.SetItemStatus(content.Public)
		}
	} else {
		return "", errors.New("content type does not implement Statusable")
	}

	b, err := c.Marshal(ci)
	if err != nil {
		return "", err
	}

	if err := c.Repo.NewContent(ci, b); err != nil {
		return "", err
	}

	return strconv.FormatInt(int64(cii.ItemID()), 10), nil
}

func (c *Content) Marshal(content any) ([]byte, error) {
	j, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (c *Content) Unmarshal(data []byte, content any) error {
	return json.Unmarshal(data, content)
}

func (c *Content) AllContentTypes() map[string]content.Creator {
	return c.Types
}

func (c *Content) NormalizeString(s string) (string, error) {
	return stringToSlug(s)
}
