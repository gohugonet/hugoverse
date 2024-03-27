package escaper

type context struct {
	state state
}

type state uint8

const (
	// stateText is parsed character data. An HTML parser is in
	// this state when its parse position is outside an HTML tag,
	// directive, comment, and special element body.
	stateText state = iota
	// stateTag occurs before an HTML attribute or the end of a tag.
	stateTag
	// stateAttrName occurs inside an attribute name.
	// It occurs between the ^'s in ` ^name^ = value`.
	stateAttrName
)
