package repository

import "net/url"

type Repository interface {
	PutConfig(key string, value any) error
	SetConfig(data url.Values) error
	LoadConfig() ([]byte, error)
	User(email string) ([]byte, error)
	PutUser(email string, data []byte) error
	NextUserId(email string) (uint64, error)
}
