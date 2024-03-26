package hstring

type RenderedString string

func (s RenderedString) String() string {
	return string(s)
}
