package contenthub

type ThemeProvider interface {
	Name() string
}

const (
	KindPage    = "page"
	KindHome    = "home"
	KindSection = "section"
)

// ResourceProvider is used to create and refresh, and clone resources needed.
type ResourceProvider interface {
	Update() error
	Clone() error
}
