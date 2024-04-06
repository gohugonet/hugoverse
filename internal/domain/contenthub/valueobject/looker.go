package valueobject

type LayoutLooker struct {
	names  []string
	bnames []string
}

func NewLayoutLooker(names []string, bnames []string) *LayoutLooker {
	return &LayoutLooker{
		names:  names,
		bnames: bnames,
	}
}

func (l *LayoutLooker) Names() []string {
	return l.names
}

func (l *LayoutLooker) BaseNames() []string {
	return l.bnames
}
