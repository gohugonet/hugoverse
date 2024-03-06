package admin

import "net/url"

type Admin interface {
	Name() string

	Editor
	UserService
	Persistence
	Cache
	Config
	Http
	Client
}

type Editor interface {
	ConfigEditor() ([]byte, error)
}

type User interface {
	Name() string
}

type UserService interface {
	ValidateUser(email, password string) error
	NewUser(email, password string) (User, error)
}

type Persistence interface {
	PutConfig(key string, value any) error
	SetConfig(data url.Values) error
}

type Http interface {
	Domain() string
	HttpPort() string
}

type Cache interface {
	InvalidateCache() error
	ETage() string
	NewETage() string
	CacheMaxAge() int64
}

type Config interface {
	CacheDisabled() bool
	CorsDisabled() bool
	GzipDisabled() bool
}

type Client interface {
	ClientSecret() string
}
