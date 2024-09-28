package db

import (
	bolt "go.etcd.io/bbolt"
	"log"
	"path/filepath"
)

var (
	store *bolt.DB

	// TODO: move it out of pkg/db
	buckets = []string{
		"__config", "__users",
		"__addons", "__uploads",
		"__contentIndex",
	}

	bucketsToAdd []string
)

// Start creates a db connection, initializes db with required info, sets secrets
func Start(dataDir string, contentTypes []string) {
	if store != nil {
		return
	}

	var err error
	systemDb := filepath.Join(dataDir, "system.db")
	store, err = bolt.Open(systemDb, 0666, nil)
	if err != nil {
		log.Fatalln("Couldn't open db.", err)
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
		}

		// init db with other buckets as needed
		buckets = append(buckets, bucketsToAdd...)

		for _, name := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalln("Couldn't initialize db with buckets.", err)
	}
}

// Close exports the abillity to close our db file. Should be called with defer
// after call to Init() from the same place.
func Close() {
	err := store.Close()
	if err != nil {
		log.Println(err)
	}
}

// SystemInitComplete checks if there is at least 1 admin user in the db which
// would indicate that the system has been configured to the minimum required.
func SystemInitComplete() bool {
	complete := false

	err := store.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("__users"))
		if users == nil {
			return bolt.ErrBucketNotFound
		}

		err := users.ForEach(func(k, v []byte) error {
			complete = true
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		complete = false
		log.Fatalln(err)
	}

	return complete
}
