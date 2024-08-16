package valueobject

type Compiler struct {
	ver string
}

func NewVersion(ver string) *Compiler {
	return &Compiler{ver: ver}
}

func (v Compiler) Version() string {
	return v.ver
}
