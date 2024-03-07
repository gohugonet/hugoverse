package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

func CheckSlugForDuplicate(slug string) (string, error) {
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

func SetIndex(id string, item Item) error {
	var err error
	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__contentIndex"))
		if err != nil {
			return err
		}

		k := []byte(item.Namespace())
		v := []byte(fmt.Sprintf("%s:%s", item.Bucket(), id))

		err = b.Put(k, v)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
