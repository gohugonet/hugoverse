package repository

type Repository interface {
	PutContent(id, slug, ns, status string, data []byte) error
	NextContentId(ns string) (uint64, error)
	CheckSlugForDuplicate(slug string) (string, error)
}
