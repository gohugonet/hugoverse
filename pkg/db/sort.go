package db

import (
	bolt "go.etcd.io/bbolt"
	"log"
)

func (s *Store) Sort(items Items) error {
	// store in <namespace>_sorted bucket, first delete existing
	err := s.db.Update(func(tx *bolt.Tx) error {
		bn := []byte(items.Bucket())
		err := tx.DeleteBucket(bn)
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		b, err := tx.CreateBucketIfNotExists(bn)
		if err != nil {
			return err
		}

		// encode to json and store as 'post.Time():i':post
		for _, kv := range items.KeyValues() {
			err = b.Put([]byte(kv.Key()), kv.Value())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Println("Error while updating db with sorted", items.Bucket(), err)
		return err
	}

	return nil
}
