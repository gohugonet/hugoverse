package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	bolt "go.etcd.io/bbolt"
	"net/url"
	"strings"
)

const ItemBucketPrefix = "__"

// ItemAll gets the configuration from the db
func ItemAll(item Item) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketNameWithPrefix(item.Bucket())))
		if b == nil {
			return fmt.Errorf("error finding bucket: %s", item.Bucket())
		}
		_, err := val.Write(b.Get([]byte(item.Namespace())))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

func bucketNameWithPrefix(name string) string {
	return ItemBucketPrefix + name
}

func marshalItem(item Item) ([]byte, error) {
	data, err := json.Marshal(item.Object())
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetItem sets key:value pairs in the db for item object
func SetItem(data url.Values, item Item) error {
	var j []byte
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketNameWithPrefix(item.Bucket())))

		// check for any multi-value fields (ex. checkbox fields)
		// and correctly format for db storage. Essentially, we need
		// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
		fieldOrderValue := make(map[string]map[string][]string)
		for k, v := range data {
			if strings.Contains(k, ".") {
				fo := strings.Split(k, ".")

				// put the order and the field value into map
				field := string(fo[0])
				order := string(fo[1])
				if len(fieldOrderValue[field]) == 0 {
					fieldOrderValue[field] = make(map[string][]string)
				}

				// orderValue is 0:[?type=Thing&id=1]
				orderValue := fieldOrderValue[field]
				orderValue[order] = v
				fieldOrderValue[field] = orderValue

				// discard the post form value with name.N
				data.Del(k)
			}

		}

		// add/set the key & value to the post form in order
		for f, ov := range fieldOrderValue {
			for i := 0; i < len(ov); i++ {
				position := fmt.Sprintf("%d", i)
				fieldValue := ov[position]

				if data.Get(f) == "" {
					for i, fv := range fieldValue {
						if i == 0 {
							data.Set(f, fv)
						} else {
							data.Add(f, fv)
						}
					}
				} else {
					for _, fv := range fieldValue {
						data.Add(f, fv)
					}
				}
			}
		}

		cfg := item.Object()
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err := dec.Decode(cfg, data)
		if err != nil {
			return err
		}

		// check for "invalidate" value to reset the Etag
		if item.CacheInvalidate() {
			//cfg.Etag = NewEtag()
			//cfg.CacheInvalidate = []string{}
		}

		j, err = json.Marshal(cfg)
		if err != nil {
			return err
		}

		err = b.Put([]byte("settings"), j)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// convert json => map[string]interface{}
	var kv map[string]interface{}
	err = json.Unmarshal(j, &kv)
	if err != nil {
		return err
	}

	mu.Lock()
	configCache = kv
	mu.Unlock()

	return nil
}
