package valueobject

type LinkNode struct {
	text string
	url  string
}

func (h *LinkNode) Text() string {
	return h.text
}

func (h *LinkNode) URL() string {
	return h.url
}
