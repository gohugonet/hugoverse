package db

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/schema"
	bolt "go.etcd.io/bbolt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SetUpload stores information about files uploaded to the system
func SetUpload(target string, data url.Values) (int, error) {
	parts := strings.Split(target, ":")
	if parts[0] != "__uploads" {
		return 0, fmt.Errorf("cannot call SetUpload with target type: %s", parts[0])
	}
	pid := parts[1]

	if data.Get("uuid") == "" ||
		data.Get("uuid") == (uuid.UUID{}).String() {

		// set new UUID for upload
		uid, err := uuid.NewV4()
		if err != nil {
			return 0, err
		}
		data.Set("uuid", uid.String())
	}

	if data.Get("slug") == "" {
		// create slug based on filename and timestamp/updated fields
		slug := data.Get("name")
		slug, err := checkSlugForDuplicate(slug)
		if err != nil {
			return 0, err
		}
		data.Set("slug", slug)
	}

	ts := fmt.Sprintf("%d", time.Now().Unix()*1000)
	if data.Get("timestamp") == "" {
		data.Set("timestamp", ts)
	}

	data.Set("updated", ts)

	// store in database
	var id uint64
	var err error
	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__uploads"))
		if err != nil {
			return err
		}

		if pid == "-1" {
			// get sequential ID for item
			id, err = b.NextSequence()
			if err != nil {
				return err
			}
			data.Set("id", fmt.Sprintf("%d", id))
		} else {
			uid, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return err
			}
			id = uint64(uid)
			data.Set("id", fmt.Sprintf("%d", id))
		}

		// todo file upload
		file := &item.FileUpload{}
		dec := schema.NewDecoder()
		dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
		dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
		err = dec.Decode(file, data)
		if err != nil {
			return err
		}

		// marshal data to json for storage
		j, err := json.Marshal(file)
		if err != nil {
			return err
		}

		uploadKey, err := key(data.Get("id"))
		if err != nil {
			return err
		}
		err = b.Put(uploadKey, j)
		if err != nil {
			return err
		}

		// add slug to __contentIndex for lookup
		b, err = tx.CreateBucketIfNotExists([]byte("__contentIndex"))
		if err != nil {
			return err
		}

		k := []byte(data.Get("slug"))
		v := []byte(fmt.Sprintf("__uploads:%d", id))

		err = b.Put(k, v)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func checkSlugForDuplicate(slug string) (string, error) {
	// check for existing slug in __contentIndex
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__contentIndex"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		original := slug
		exists := true
		i := 0
		for exists {
			s := b.Get([]byte(slug))
			if s == nil {
				exists = false
				return nil
			}

			i++
			slug = fmt.Sprintf("%s-%d", original, i)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return slug, nil
}

func key(sid string) ([]byte, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b, err
}
