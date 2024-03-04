package repository

type Repository interface {
	PutConfig(key string, value any) error
}
