package db

type Item interface {
	Bucket() string    // bucket name
	Namespace() string // bucket field
	Object() any       // Object instance
}
