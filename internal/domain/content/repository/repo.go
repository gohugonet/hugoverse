package repository

type Repository interface {
	PutContent(ci any, data []byte) error
	NewContent(ci any, data []byte) error

	GetContent(namespace string, id string) ([]byte, error)
	DeleteContent(namespace string, id string, slug string) error

	NextContentId(ns string) (uint64, error)
	CheckSlugForDuplicate(slug string) (string, error)
}
