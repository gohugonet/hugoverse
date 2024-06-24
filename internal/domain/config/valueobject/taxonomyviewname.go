package valueobject

type ViewName struct {
	Singular      string // e.g. "category"
	Plural        string // e.g. "categories"
	PluralTreeKey string
}

func (v ViewName) IsZero() bool {
	return v.Singular == ""
}
