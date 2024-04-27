package urls

type URL interface {
	RelURL(in string) string
	AbsURL(in string) string
	URLize(uri string) string
}

type RefLinker interface {
	Ref(args map[string]any) (string, error)
	RelRef(args map[string]any) (string, error)
}
