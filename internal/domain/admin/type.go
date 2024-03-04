package admin

type Admin interface {
	PutConfig(key string, value any) error
}
