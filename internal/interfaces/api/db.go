package api

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/search"
	"github.com/gohugonet/hugoverse/pkg/db"
	"log"
	"strconv"
)

type database struct{}

func (d *database) start(contentTypeNames []string) {
	db.Start(dataDir(), contentTypeNames)
}

func (d *database) close() {
	db.Close()
}

func (d *database) User(email string) ([]byte, error) {
	return db.Get(newUserItem(email, nil))
}

func (d *database) PutUser(email string, data []byte) error {
	return db.Set(newUserItem(email, data))
}

func (d *database) PutSortedContent(namespace string, m map[string][]byte) error {
	return db.Sort(newItems(namespace, m))
}

func (d *database) AllContent(namespace string) [][]byte {
	return db.ContentAll(namespace)
}

func (d *database) GetContent(contentType string, id string) ([]byte, error) {
	return db.Get(
		&item{
			bucket: contentType,
			key:    id,
		})
}

func (d *database) DeleteContent(namespace string, id string, slug string) error {
	if err := db.Delete(&item{bucket: namespace, key: id}); err != nil {
		return err
	}

	if err := db.RemoveIndex(slug); err != nil {
		return err
	}

	return nil
}

func (d *database) PutContent(ci any, data []byte) error {
	cii, ok := ci.(content.Identifiable)
	if !ok {
		return errors.New("invalid content type")
	}
	id := cii.ItemID()
	ns := cii.ItemName()

	cis, ok := ci.(content.Statusable)
	if !ok {
		return errors.New("invalid content type")
	}
	status := cis.ItemStatus()

	bucket := ns
	if !(status == content.Public || status == "") {
		bucket = fmt.Sprintf("%s%s", ns, bucketNameWithPrefix(string(status)))
	}

	fmt.Printf(" === bucket: %s\n", bucket)

	if err := db.Set(
		&item{
			bucket: bucket,
			key:    strconv.FormatInt(int64(id), 10),
			value:  data,
		}); err != nil {
		return err
	}

	go func() {
		// update data in search index
		if err := search.UpdateIndex(ns, fmt.Sprintf("%d", id), data); err != nil {
			log.Println("[search] UpdateIndex Error:", err)
		}
	}()

	return nil
}

func (d *database) NewContent(ci any, data []byte) error {
	if err := d.PutContent(ci, data); err != nil {
		return err
	}

	cii, ok := ci.(content.Identifiable)
	if !ok {
		return errors.New("invalid content type")
	}
	id := cii.ItemID()
	ns := cii.ItemName()

	ciSlug, ok := ci.(content.Sluggable)
	if !ok {
		return errors.New("invalid content type")
	}
	if err := db.SetIndex(newKeyValueItem(ciSlug.ItemSlug(), fmt.Sprintf("%s:%d", ns, id))); err != nil {
		return err
	}

	return nil
}

func (d *database) NextContentId(ns string) (uint64, error) {
	return db.NextSequence(&item{bucket: ns})
}

func (d *database) NextUserId(email string) (uint64, error) {
	return db.NextSequence(newBucketItem("users"))
}

func (d *database) NextUploadId() (uint64, error) {
	return db.NextSequence(newBucketItem("uploads"))
}

func (d *database) GetUpload(id string) ([]byte, error) {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return nil, err
	}
	return db.Get(newUploadItem(key, nil))
}

func (d *database) DeleteUpload(id string) error {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return err
	}
	return db.Delete(newUploadItem(key, nil))
}

func (d *database) AllUploads() ([][]byte, error) {
	return db.All(newUploadItem("", nil))
}

func (d *database) NewUpload(id, slug string, data []byte) error {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return err
	}
	if err := db.Set(newUploadItem(key, data)); err != nil {
		return err
	}

	if err := db.SetIndex(newKeyValueItem(
		slug,
		fmt.Sprintf("%s:%s", bucketNameWithPrefix("upload"), id))); err != nil {
		return err
	}

	return nil
}

func keyBit8Uint64(sid string) (string, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return "", err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return string(b), err
}

func (d *database) PutConfig(data []byte) error {
	return db.Set(newConfigItem(data))
}

func (d *database) LoadConfig() ([]byte, error) {
	data, err := db.Get(newConfigItem(nil))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *database) CheckSlugForDuplicate(slug string) (string, error) {
	return db.CheckSlugForDuplicate(slug)
}
