package api

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/pkg/db"
	"net/url"
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
	var user *entity.User
	if err := json.Unmarshal(data, &user); err != nil {
		return err
	}
	return db.Set(newUserItem(email, user))
}

func (d *database) NextUserId(email string) (uint64, error) {
	return db.NextSequence(newUserItem(email, nil))
}

func (d *database) SetConfig(data url.Values) error {
	return db.SetItem(data, newConfigItem(&entity.Config{}))
}

func (d *database) PutConfig(key string, value interface{}) error {
	return db.PutItem(key, value, newConfigItem(&entity.Config{}))
}

func (d *database) LoadConfig() ([]byte, error) {
	data, err := db.ItemAll(newConfigItem(&entity.Config{}))
	if err != nil {
		return nil, err
	}

	return data, nil
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

func newConfigItem(object *entity.Config) *item {
	return &item{
		bucket: bucketNameWithPrefix("config"),
		ns:     "settings",
		object: object,
	}
}

func newUserItem(email string, user *entity.User) *item {
	return &item{
		bucket: bucketNameWithPrefix("users"),
		ns:     email,
		object: user,
	}
}

type item struct {
	bucket string
	ns     string
	object any
}

func (it *item) Bucket() string    { return it.bucket }
func (it *item) Namespace() string { return it.ns }
func (it *item) Object() any       { return it.object }
