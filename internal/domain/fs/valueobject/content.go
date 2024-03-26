package valueobject

type ContentClass string

const (
	ContentClassLeaf    ContentClass = "leaf"
	ContentClassBranch  ContentClass = "branch"
	ContentClassFile    ContentClass = "zfile" // Sort below
	ContentClassContent ContentClass = "zcontent"
)
