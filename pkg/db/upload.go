package db

import (
	"encoding/binary"
	bolt "go.etcd.io/bbolt"
	"strconv"
)

// SetUpload stores information about files uploaded to the system
func SetUpload(data []byte, item Item) error {
	var err error
	err = store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(item.Bucket()))
		if err != nil {
			return err
		}

		uploadKey, err := key(item.Namespace())
		if err != nil {
			return err
		}
		err = b.Put(uploadKey, data)
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

func key(sid string) ([]byte, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b, err
}
