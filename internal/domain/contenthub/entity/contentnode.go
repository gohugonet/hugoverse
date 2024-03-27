package entity

import (
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"strings"
)

type contentNode struct {
	p *pageState

	// Set for taxonomy nodes.
	//viewInfo *contentBundleViewInfo

	// Set if source is a file.
	// We will soon get other sources.
	fi fsVO.FileMetaInfo

	// The source path. Unix slashes. No leading slash.
	path string
}

func (b *contentNode) rootSection() string {
	if b.path == "" {
		return ""
	}
	firstSlash := strings.Index(b.path, "/")
	if firstSlash == -1 {
		return b.path
	}
	return b.path[:firstSlash]
}
