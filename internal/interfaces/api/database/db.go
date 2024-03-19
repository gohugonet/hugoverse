package database

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

type Database struct {
	dataDir string
}

func New(dataDir string) *Database {
	return &Database{
		dataDir: dataDir,
	}
}

func (d *Database) Start(contentTypeNames []string) {
	db.Start(d.dataDir, contentTypeNames)
}

func (d *Database) Close() {
	db.Close()
}

func (d *Database) User(email string) ([]byte, error) {
	return db.Get(newUserItem(email, nil))
}

func (d *Database) PutUser(email string, data []byte) error {
	return db.Set(newUserItem(email, data))
}

func (d *Database) PutSortedContent(namespace string, m map[string][]byte) error {
	return db.Sort(newItems(namespace, m))
}

func (d *Database) AllContent(namespace string) [][]byte {
	return db.ContentAll(namespace)
}

func (d *Database) GetContent(contentType string, id string) ([]byte, error) {
	return db.Get(
		&item{
			bucket: contentType,
			key:    id,
		})
}

func (d *Database) DeleteContent(namespace string, id string, slug string) error {
	if err := db.Delete(&item{bucket: namespace, key: id}); err != nil {
		return err
	}

	if err := db.RemoveIndex(slug); err != nil {
		return err
	}

	return nil
}

func (d *Database) PutContent(ci any, data []byte) error {
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

func (d *Database) NewContent(ci any, data []byte) error {
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

func (d *Database) NextContentId(ns string) (uint64, error) {
	return db.NextSequence(&item{bucket: ns})
}

func (d *Database) NextUserId(email string) (uint64, error) {
	return db.NextSequence(newBucketItem("users"))
}

func (d *Database) NextUploadId() (uint64, error) {
	return db.NextSequence(newBucketItem("uploads"))
}

func (d *Database) GetUpload(id string) ([]byte, error) {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return nil, err
	}
	return db.Get(newUploadItem(key, nil))
}

func (d *Database) DeleteUpload(id string) error {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return err
	}
	return db.Delete(newUploadItem(key, nil))
}

func (d *Database) AllUploads() ([][]byte, error) {
	return db.All(newUploadItem("", nil))
}

func (d *Database) NewUpload(id, slug string, data []byte) error {
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

func (d *Database) PutConfig(data []byte) error {
	return db.Set(newConfigItem(data))
}

func (d *Database) LoadConfig() ([]byte, error) {
	data, err := db.Get(newConfigItem(nil))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *Database) CheckSlugForDuplicate(slug string) (string, error) {
	return db.CheckSlugForDuplicate(slug)
}
