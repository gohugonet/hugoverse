package parser

type baseNode struct {
	val string
}

func (b *baseNode) SetVal(v string) {
	b.val = v
}
