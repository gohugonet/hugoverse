package media

// TODO, remove it

// DefaultTypes is the default media types supported by Hugo.
var DefaultTypes = Types{
	Builtin.HTMLType,
}

// DecodeTypes takes a list of media type configurations and merges those,
// in the order given, with the Hugo defaults as the last resort.
func DecodeTypes() Types {
	var m Types

	// remove duplications
	// Maps type string to Type. Type string is the full application/svg+xml.
	mmm := make(map[string]Type)
	for _, dt := range DefaultTypes {
		mmm[dt.Type] = dt
	}

	for _, v := range mmm {
		m = append(m, v)
	}

	return m
}
