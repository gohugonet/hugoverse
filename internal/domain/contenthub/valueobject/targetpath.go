package valueobject

import (
	"path"
	"path/filepath"
	"strings"
)

type TargetPaths struct {

	// Where to store the file on disk relative to the publish dir. OS slashes.
	TargetFilename string

	// The directory to write sub-resources of the above.
	SubResourceBaseTarget string

	// The base for creating links to sub-resources of the above.
	SubResourceBaseLink string

	// The relative permalink to this resources. Unix slashes.
	Link string
}

func (p TargetPaths) RelPermalink() string {
	return p.PrependBasePath(p.Link, false)
}

// PrependBasePath prepends any baseURL sub-folder to the given resource
func (p TargetPaths) PrependBasePath(rel string, isAbs bool) string {
	basePath := p.GetBasePath(!isAbs)
	if basePath != "" {
		rel = filepath.ToSlash(rel)
		// Need to prepend any path from the baseURL
		hadSlash := strings.HasSuffix(rel, "/")
		rel = path.Join(basePath, rel)
		if hadSlash {
			rel += "/"
		}
	}
	return rel
}

// GetBasePath returns any path element in baseURL if needed.
func (p TargetPaths) GetBasePath(isRelativeURL bool) string {
	return ""
}

func (p TargetPaths) PermalinkForOutputFormat() string {
	return p.PermalinkForBaseURL(p.Link, "")
}

// PermalinkForBaseURL creates a permalink from the given link and baseURL.
func (p TargetPaths) PermalinkForBaseURL(link, baseURL string) string {
	link = strings.TrimPrefix(link, "/")
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL + link
}
