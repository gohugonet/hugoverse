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

type ParseConfig struct {
	Version int
}