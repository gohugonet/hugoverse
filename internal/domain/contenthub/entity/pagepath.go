package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/paths/files"
)

type Path struct {
	pathInfo *paths.Path
}

func newPathFromConfig(path string, kind string, pi *paths.Path) *Path {
	s := path
	if !paths.HasExt(s) {
		var (
			isBranch bool
			ext      string = "md"
		)
		if kind != "" {
			isBranch = valueobject.IsBranch(kind)
		} else if pi != nil {
			isBranch = pi.IsBranchBundle()
			if pi.Ext() != "" {
				ext = pi.Ext()
			}
		}

		if isBranch {
			s += "/_index." + ext
		} else {
			s += "/index." + ext
		}
	}
	pathInfo := paths.Parse(files.ComponentFolderContent, s)

	return &Path{
		pathInfo: pathInfo,
	}
}

func (p *Path) Path() string {
	return p.pathInfo.Base()
}
