package valueobject

type ParseInfo struct {
	// Set for shortcode templates with any {{ .Inner }}
	IsInner bool

	// Set for partials with a return statement.
	HasReturn bool

	// Config extracted from template.
	Config ParseConfig
}

func (info ParseInfo) IsZero() bool {
	return info.Config.Version == 0
}
func (info ParseInfo) Return() bool {
	return info.HasReturn
}

func (info ParseInfo) Inner() bool {
	return info.IsInner
}

type ParseConfig struct {
	Version int
}
