package repository

type Repository interface {
	PutContent(ci any, data []byte) error
	NextContentId(ns string) (uint64, error)
	CheckSlugForDuplicate(slug string) (string, error)
}
