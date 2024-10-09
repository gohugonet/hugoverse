package db

import (
	bolt "go.etcd.io/bbolt"
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
