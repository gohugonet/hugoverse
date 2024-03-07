package db

type Item interface {
	BucketItem
	Namespace() string // bucket field
	Object() any       // Object instance
}

type BucketItem interface {
	Bucket() string
}
