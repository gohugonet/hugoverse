package db

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
)

var ErrExists = errors.New("exists")

func Get(item Item) ([]byte, error) {
	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(item.Bucket()))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		obj := b.Get([]byte(item.Namespace()))

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
		email := []byte(item.Namespace())
		users := tx.Bucket([]byte(item.Bucket()))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		// check if user is found by email, fail if nil
		exists := users.Get(email)
		if exists != nil {
			return ErrExists
		}

		// marshal User to json and put into bucket
		j, err := json.Marshal(item.Object())
		if err != nil {
			return err
		}

		err = users.Put(email, j)
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
