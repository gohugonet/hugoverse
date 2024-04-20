package valueobject

type Component struct {
	ComName string
	ComDir  string
	ComLang string
}

func (c *Component) Name() string {
	return c.ComName
}

func (c *Component) Dir() string {
	return c.ComDir
}

func (c *Component) Language() string {
	return c.ComLang
}
