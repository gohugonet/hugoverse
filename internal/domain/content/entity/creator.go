package entity

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/pkg/form"
	"github.com/gorilla/schema"
	"log"
	"net/url"
	"strconv"
)

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

	// TODO， need to sync content to file system
	// in hugo project way
	// check changes: file system existing check? hash?
	return c.newContent(contentType, ci)
}

func (c *Content) newContent(contentType string, ci any) (string, error) {
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

	if cis.ItemStatus() == content.Public {
		go func() {
			if err := c.SortContent(contentType); err != nil {
				log.Println("sort content err: ", err)
			}
		}()
	}

	id := int64(cii.ItemID())

	go func() {
		// update data in search index
		if err := c.Search.UpdateIndex(
			GetNamespace(contentType, string(cis.ItemStatus())),
			fmt.Sprintf("%d", id), b); err != nil {

			log.Println("[search] UpdateIndex Error:", err)
		}
	}()

	return strconv.FormatInt(id, 10), nil
}