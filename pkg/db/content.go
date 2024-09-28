package db

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
)

// ContentAll retrives all items from the database within the provided namespace
func ContentAll(namespace string) [][]byte {
	var posts [][]byte

	if err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			fmt.Println("Bucket not found", namespace)
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		posts = make([][]byte, 0, numKeys)

		if err := b.ForEach(func(k, v []byte) error {
			posts = append(posts, v)

			return nil
		}); err != nil {
			log.Println("Error reading from db", namespace, err)
			return err
		}

		return nil
	}); err != nil {
		log.Println("Error reading from db", namespace, err)
	}

	return posts
}
