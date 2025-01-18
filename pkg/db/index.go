package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) RemoveIndex(slug string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		ci := tx.Bucket([]byte("__contentIndex"))
		if ci == nil {
			return bolt.ErrBucketNotFound
		}

		err := ci.Delete([]byte(slug))
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

func (s *Store) CheckSlugForDuplicate(slug string) (string, error) {
	// check for existing slug in __contentIndex
	err := s.db.View(func(tx *bolt.Tx) error {
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

func (s *Store) GetIndex(key string) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("__contentIndex"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		value = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Store) SetIndex(item KeyValue) error {
	var err error
	err = s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("__contentIndex"))
		if err != nil {
			return err
		}

		k := []byte(item.Key())
		v := item.Value()

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
