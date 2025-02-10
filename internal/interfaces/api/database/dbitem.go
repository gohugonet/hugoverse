package database

import (
	"fmt"
	"github.com/mdfriday/hugoverse/pkg/db"
)

const ItemBucketPrefix = "__"
const ItemBucketIndexSuffix = "__index"

func bucketNameWithPrefix(name string) string {
	return ItemBucketPrefix + name
}

func bucketNameWithIndex(name string) string {
	return name + ItemBucketIndexSuffix
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

type items struct {
	bucket string
	kvs    []db.KeyValue
}

func (it *items) Bucket() string           { return it.bucket }
func (it *items) KeyValues() []db.KeyValue { return it.kvs }

func newItems(name string, m map[string][]byte) *items {
	kvs := make([]db.KeyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, newKeyValueItem(k, string(v)))
	}
	return &items{
		bucket: fmt.Sprintf("%s%s", name, bucketNameWithPrefix("sorted")),
		kvs:    kvs,
	}
}
