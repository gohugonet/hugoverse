package db

import bolt "go.etcd.io/bbolt"

func (s *Store) NextSequence(item BucketItem) (uint64, error) {
	var id uint64
	err := s.db.Update(func(tx *bolt.Tx) error {
		items := tx.Bucket([]byte(item.Bucket()))
		if items == nil {
			return bolt.ErrBucketNotFound
		}

		var err error
		id, err = items.NextSequence()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}
