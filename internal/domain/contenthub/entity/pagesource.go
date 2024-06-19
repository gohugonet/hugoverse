package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

type PageSource struct {
	fi     fs.FileMetaInfo
	path   *paths.Path
	opener io.OpenReadSeekCloser

	stale.Staler
}

func newPageSource(fi fs.FileMetaInfo) (*PageSource, error) {
	path := paths.Parse("", fi.FileName())
	r := func() (io.ReadSeekCloser, error) {
		return fi.Open()
	}
	return &PageSource{
		fi:     fi,
		path:   path,
		opener: r,

		Staler: &stale.AtomicStaler{},
	}, nil
}
