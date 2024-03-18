package db

import (
	"bytes"
	bolt "go.etcd.io/bbolt"
)

func All(item Item) ([][]byte, error) {
	var items [][]byte
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(item.Bucket()))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		items = make([][]byte, 0, numKeys)

		return b.ForEach(func(k, v []byte) error {
			items = append(items, v)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

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
		bucket, err := tx.CreateBucketIfNotExists([]byte(item.Bucket()))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(item.Key()), item.Value())
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
