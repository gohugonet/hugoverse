package db

type Item interface {
	BucketItem
	KeyValue
}

type KeyValue interface {
	Key() string
	Value() []byte
}

type BucketItem interface {
	Bucket() string
}
