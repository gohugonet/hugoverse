package valueobject

type BundleType int

const (
	// BundleTypeFile A generic resource, e.g. a JSON file.
	BundleTypeFile BundleType = iota

	// BundleTypeContentResource All below are content files.
	// A resource of a content type with front matter.
	// A single file but not leaf
	BundleTypeContentResource

	// BundleTypeContentSingle E.g. /blog/my-post.md
	BundleTypeContentSingle

	// All below are bundled content files.

	// BundleTypeLeaf Leaf bundles, e.g. /blog/my-post/index.md
	BundleTypeLeaf

	// BundleTypeBranch Branch bundles, e.g. /blog/_index.md
	BundleTypeBranch
)

func (b BundleType) IsBundle() bool {
	return b >= BundleTypeLeaf
}

func (b BundleType) IsLeafBundle() bool {
	return b == BundleTypeLeaf
}

func (b BundleType) IsBranchBundle() bool {
	return b == BundleTypeBranch
}

func (b BundleType) IsContentResource() bool {
	return b == BundleTypeContentResource
}

// IsContent returns true if the path is a content file (e.g. mypost.md).
// Note that this will also return true for content files in a bundle.
func (b BundleType) IsContent() bool {
	return b >= BundleTypeContentResource
}
