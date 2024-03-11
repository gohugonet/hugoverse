package api

import (
	"encoding/binary"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/db"
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

func (d *database) PutContent(id, slug, ns, status string, data []byte) error {
	if err := db.Set(newContentItem(id, ns, status, data)); err != nil {
		return err
	}

	if err := db.SetIndex(newKeyValueItem(
		slug, fmt.Sprintf("%s:%s", ns, id))); err != nil {
		return err
	}

	if status == "" {
		go db.SortContent(ns)
	}

	return nil
}

func (d *database) NextContentId(ns string) (uint64, error) {
	return db.NextSequence(newBucketItem(ns))
}

func (d *database) NextUserId(email string) (uint64, error) {
	return db.NextSequence(newBucketItem("users"))
}

func (d *database) NextUploadId() (uint64, error) {
	return db.NextSequence(newBucketItem("uploads"))
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

const ItemBucketPrefix = "__"

func bucketNameWithPrefix(name string) string {
	return ItemBucketPrefix + name
}

func newUploadItem(id string, data []byte) *item {
	return &item{
		bucket: bucketNameWithPrefix("uploads"),
		key:    id,
		value:  data,
	}
}

func newConfigItem(val []byte) *item {
	return &item{
		bucket: bucketNameWithPrefix("config"),
		key:    "settings",
		value:  val,
	}
}

func newUserItem(email string, user []byte) *item {
	return &item{
		bucket: bucketNameWithPrefix("users"),
		key:    email,
		value:  user,
	}
}

func newContentItem(id, ns, status string, data []byte) *item {
	return &item{
		bucket: fmt.Sprintf("%s%s", ns, bucketNameWithPrefix(status)),
		key:    id,
		value:  data,
	}
}

func newKeyValueItem(key, value string) *item {
	return &item{
		key:   key,
		value: []byte(value),
	}
}

func newBucketItem(name string) *item {
	return &item{
		bucket: bucketNameWithPrefix(name),
	}
}

type item struct {
	bucket string
	key    string
	value  []byte
}

func (it *item) Bucket() string { return it.bucket }
func (it *item) Key() string    { return it.key }
func (it *item) Value() []byte  { return it.value }
