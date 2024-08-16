package site

type Service interface {
	Author
}

type Author interface {
	Name() string
	Email() string
}
