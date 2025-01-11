package db

import (
	"github.com/gohugonet/hugoverse/pkg/loggers"
	bolt "go.etcd.io/bbolt"
	"path/filepath"
)

type Store struct {
	db  *bolt.DB
	log loggers.Logger
}

func (s *Store) Close() error {
	return s.db.Close()
}

func NewStore(dataDir string, contentTypes []string) (*Store, error) {
	log := loggers.NewDefault()

	var err error
	systemDb := filepath.Join(dataDir, "system.db")
	store, err = bolt.Open(systemDb, 0666, nil)
	if err != nil {
		log.Errorln("Couldn't open db.", err)
		return nil, err
	}

	err = store.Update(func(tx *bolt.Tx) error {
		// initialize db with all content type buckets & sorted bucket for type
		for _, t := range contentTypes {
			_, err := tx.CreateBucketIfNotExists([]byte(t))
			if err != nil {
				return err
			}

			_, err = tx.CreateBucketIfNotExists([]byte(t + "__sorted"))
			if err != nil {
				return err
			}

			_, err = tx.CreateBucketIfNotExists([]byte(t + "__index"))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Errorln("Couldn't initialize db with buckets.", err)
		return nil, err
	}

	return &Store{store, log}, nil
}
