package db

import (
	"bytes"
	bolt "go.etcd.io/bbolt"
)

func Get(item Item) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(item.Bucket()))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		obj := b.Get([]byte(item.Key()))

		_, err := val.Write(obj)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if val.Bytes() == nil {
		return nil, nil
	}

	return val.Bytes(), nil
}

func Set(item Item) error {
	err := store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(item.Bucket()))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		err := bucket.Put([]byte(item.Key()), item.Value())
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

func Delete(item Item) error {
	err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(item.Bucket()))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(item.Key()))
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
