package db

import (
	"bytes"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
)

// ContentAll retrives all items from the database within the provided namespace
func (s *Store) ContentAll(namespace string) [][]byte {
	var posts [][]byte

	if err := s.db.View(func(tx *bolt.Tx) error {
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

// ContentByPrefix retrieves all raw byte items from the database with a specific slug prefix.
func (s *Store) ContentByPrefix(namespace, prefix string) ([][]byte, error) {
	var results [][]byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			fmt.Println("Bucket not found:", namespace)
			return bolt.ErrBucketNotFound
		}

		c := b.Cursor()

		// 定位到第一个匹配 slugPrefix 的键
		for k, v := c.Seek([]byte(prefix)); k != nil && bytes.HasPrefix(k, []byte(prefix)); k, v = c.Next() {
			results = append(results, v)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error reading from db with slug prefix:", namespace, err)
		return nil, err
	}

	return results, nil
}
