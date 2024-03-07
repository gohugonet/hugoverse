package db

import (
	"bytes"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/url"
	"strings"
)

// Content retrives one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func Content(target string) ([]byte, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := val.Write(b.Get([]byte(id)))
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// SetContent inserts/replaces values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func SetContent(target string, data url.Values) (int, error) {
	//t := strings.Split(target, ":")
	//ns, id := t[0], t[1]

	// check if content id == -1 (indicating new post).
	// if so, run an insert which will assign the next auto incremented int.
	// this is done because bbolt begins its bucket auto increment value at 0,
	// which is the zero-value of an int in the Item struct field for ID.
	// this is a problem when the original first post (with auto ID = 0) gets
	// overwritten by any new post, originally having no ID, defauting to 0.
	//if id == "-1" {
	//	return insert(ns, data)
	//}

	return 0, nil

	//return update(ns, id, data, nil)
}
