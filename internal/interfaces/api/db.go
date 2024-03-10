package api

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/pkg/db"
)

type database struct {
}

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

func (d *database) NextUserId(email string) (uint64, error) {
	return db.NextSequence(newUserItem(email, nil))
}

func (d *database) NextUploadId() (uint64, error) {
	return db.NextSequence(newUploadBucketItem())
}

func (d *database) NewUpload(id, slug string, data []byte) error {
	if err := db.SetUpload(data, newUploadItem(id)); err != nil {
		return err
	}

	if err := db.SetIndex(id, newUploadItem(slug)); err != nil {
		return err
	}

	return nil
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

func (d *database) CheckUploadDuplication(slug string) (string, error) {
	return db.CheckSlugForDuplicate(slug)
}

func emptyConfig() ([]byte, error) {
	cfg := &entity.Config{}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}

const ItemBucketPrefix = "__"

func bucketNameWithPrefix(name string) string {
	return ItemBucketPrefix + name
}

func newUploadItem(id string) *item {
	return &item{
		bucket: bucketNameWithPrefix("uploads"),
		key:    id,
		value:  nil,
	}
}

func newUploadBucketItem() *item {
	return &item{
		bucket: bucketNameWithPrefix("uploads"),
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

type item struct {
	bucket string
	key    string
	value  []byte
}

func (it *item) Bucket() string { return it.bucket }
func (it *item) Key() string    { return it.key }
func (it *item) Value() []byte  { return it.value }
