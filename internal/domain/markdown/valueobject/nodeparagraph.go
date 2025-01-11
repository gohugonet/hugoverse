package valueobject

type ParagraphNode struct {
	text string
}

func (p *ParagraphNode) Text() string {
	return p.text
}
