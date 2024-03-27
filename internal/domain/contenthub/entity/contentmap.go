package entity

import (
	"fmt"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"path"
	"path/filepath"
	"strings"
)

type ContentMap struct {
	// View of regular pages, sections, and taxonomies.
	PageTrees ContentTrees

	// View of pages, sections, taxonomies, and resources.
	BundleTrees ContentTrees

	// Stores page bundles keyed by its path's directory or the base filename,
	// e.g. "blog/post.md" => "/blog/post", "blog/post/index.md" => "/blog/post"
	// These are the "regular pages" and all of them are bundles.
	Pages *ContentTree

	// Section nodes.
	Sections *ContentTree

	// Resources stored per bundle below a common prefix, e.g. "/blog/post__hb_".
	//Resources *ContentTree
}

func (m *ContentMap) AddFilesBundle(header fsVO.FileMetaInfo) error {
	var (
		meta       = header.Meta()
		bundlePath = m.getBundleDir(meta)

		n = m.newContentNodeFromFi(header)
		b = m.newKeyBuilder()

		section string
	)

	// A regular page. Attach it to its section.
	section, _ = m.getOrCreateSection(n, bundlePath) // /abc/
	b = b.WithSection(section).ForPage(bundlePath).Insert(n)

	fmt.Println("<<< AddFilesBundle section: ", section, bundlePath, n.p)

	return nil
}

func (m *ContentMap) getBundleDir(meta *fsVO.FileMeta) string {
	dir := cleanTreeKey(filepath.Dir(meta.Path))
	fmt.Println(">>> getBundleDir dir: ", dir)
	switch meta.Classifier {
	case fsVO.ContentClassContent:
		meta.TranslationBaseName = "post"
		fmt.Println(">>> getBundleDir111 dir: ", path.Join(dir, meta.TranslationBaseName))
		return path.Join(dir, meta.TranslationBaseName)
	default:
		return dir
	}
}

func (m *ContentMap) newContentNodeFromFi(fi fsVO.FileMetaInfo) *contentNode {
	return &contentNode{
		fi:   fi,
		path: strings.TrimPrefix(filepath.ToSlash(fi.Meta().Path), "/"),
	}
}

func (m *ContentMap) newKeyBuilder() *cmInsertKeyBuilder {
	return &cmInsertKeyBuilder{m: m}
}

func (m *ContentMap) getOrCreateSection(n *contentNode, s string) (string, *contentNode) {
	k, b := m.getSection(s)

	k = cleanSectionTreeKey(s[:strings.Index(s[1:], "/")+1])

	b = &contentNode{
		path: n.rootSection(),
	}

	m.Sections.Insert(k, b)

	return k, b
}

func (m *ContentMap) getSection(s string) (string, *contentNode) {
	s = AddTrailingSlash(path.Dir(strings.TrimSuffix(s, "/")))

	v, found := m.Sections.Get(s)
	if found {
		return s, v.(*contentNode)
	}
	return "", nil
}

// AddTrailingSlash adds a trailing Unix styled slash (/) if not already
// there.
func AddTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func cleanSectionTreeKey(k string) string {
	k = cleanTreeKey(k)
	if k != "/" {
		k += "/"
	}

	return k
}

func cleanTreeKey(k string) string {
	k = "/" + strings.ToLower(strings.Trim(path.Clean(filepath.ToSlash(k)), "./"))
	return k
}

type cmInsertKeyBuilder struct {
	m *ContentMap

	err error

	// Builder state
	tree    *ContentTree
	baseKey string // Section or page key
	key     string
}

func (b *cmInsertKeyBuilder) WithSection(s string) *cmInsertKeyBuilder {
	s = cleanSectionTreeKey(s)
	b.newTopLevel()
	b.tree = b.m.Sections
	b.baseKey = s
	b.key = s
	return b
}

func (b *cmInsertKeyBuilder) newTopLevel() {
	b.key = ""
}

const (
	cmBranchSeparator = "__hb_"
	cmLeafSeparator   = "__hl_"
)

func (b *cmInsertKeyBuilder) ForPage(s string) *cmInsertKeyBuilder {
	baseKey := b.baseKey
	b.baseKey = s

	if baseKey != "/" {
		// Don't repeat the section path in the key.
		s = strings.TrimPrefix(s, baseKey)
	}
	s = strings.TrimPrefix(s, "/")

	switch b.tree {
	case b.m.Sections:
		b.tree = b.m.Pages
		b.key = baseKey + cmBranchSeparator + s + cmLeafSeparator
	default:
		panic("invalid state")
	}

	return b
}

func (b *cmInsertKeyBuilder) Insert(n *contentNode) *cmInsertKeyBuilder {
	if b.err == nil {
		b.tree.Insert(b.Key(), n)
	}
	return b
}

func (b *cmInsertKeyBuilder) Key() string {
	switch b.tree {
	case b.m.Sections:
		return cleanSectionTreeKey(b.key)
	default:
		return cleanTreeKey(b.key)
	}
}

// Assemble

func (m *ContentMap) CreateMissingNodes() error {
	// Create missing home and root sections
	rootSections := make(map[string]any)
	rootSections["/"] = true // not found in both sections and pages

	for sect := range rootSections {
		var sectionPath string
		sect = cleanSectionTreeKey(sect)

		_, found := m.Sections.Get(sect)
		if !found {
			mm := &contentNode{path: sectionPath} // ""
			_, _ = m.Sections.Insert(sect, mm)    // "/"
		}
	}

	return nil
}

func (m *PageMap) splitKey(k string) []string {
	if k == "" || k == "/" {
		return nil
	}

	return strings.Split(k, "/")[1:]
}

// withEveryBundlePage applies fn to every Page, including those bundled inside
// leaf bundles.
func (m *PageMap) withEveryBundlePage(fn func(p *pageState) bool) {
	m.BundleTrees.Walk(func(s string, n *contentNode) bool {
		if n.p != nil {
			return fn(n.p)
		}
		return false
	})
}
