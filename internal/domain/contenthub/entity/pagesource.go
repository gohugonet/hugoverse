package entity

import (
	"bytes"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/text"
	sio "io"
	"path/filepath"
)

type PageSource struct {
	fi     fs.FileMetaInfo
	path   *paths.Path
	opener io.OpenReadSeekCloser

	contenthub.File

	// Returns the position in bytes after any front matter.
	posMainContent int
	parseInfo      *valueobject.SourceParseInfo

	stale.Staler
	cache *valueobject.Cache
}

func newPageSource(fi fs.FileMetaInfo, c *valueobject.Cache) (*PageSource, error) {
	path := paths.Parse("", fi.FileName())
	r := func() (io.ReadSeekCloser, error) {
		return fi.Open()
	}
	return &PageSource{
		fi:     fi,
		path:   path,
		opener: r,

		File: valueobject.NewFileInfo(fi),

		Staler: &stale.AtomicStaler{},
		cache:  c,
	}, nil
}

func (p *PageSource) sourceKey() string {
	return filepath.ToSlash(p.File.Filename())
}

func (p *PageSource) contentSource() ([]byte, error) {
	key := p.sourceKey()
	v, err := p.cache.CacheContentSource.GetOrCreate(key, func(string) (*stale.Value[[]byte], error) {
		b, err := p.readSourceAll()
		if err != nil {
			return nil, err
		}

		return &stale.Value[[]byte]{
			Value: b,
			IsStaleFunc: func() bool {
				return p.Staler.IsStale()
			},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return v.Value, nil
}

func (p *PageSource) readSourceAll() ([]byte, error) {
	r, err := p.fi.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return sio.ReadAll(r)
}

func (p *PageSource) parse() error {
	content, err := p.contentSource()
	if err != nil {
		return err
	}

	items, err := pageparser.ParseBytes(
		content,
		pageparser.Config{},
	)
	if err != nil {
		return err
	}

	p.parseInfo = &valueobject.SourceParseInfo{
		ItemsStep1: items,
	}
	return nil
}

type itemMap func(item pageparser.Item, source []byte) error

func (p *PageSource) mapItems(frontMatter itemMap) error {
	if p.parseInfo.IsEmpty() {
		return nil
	}

	source, err := p.contentSource()
	if err != nil {
		return err
	}

	iter := pageparser.NewIterator(p.parseInfo.ItemsStep1)

Loop:
	for {
		it := iter.Next()
		switch {
		case it.IsFrontMatter():
			if err := frontMatter(it, source); err != nil {
				var fe herrors.FileError
				if errors.As(err, &fe) {
					pos := fe.Position()

					// Offset the starting position of front matter.
					offset := iter.LineNumber(source) - 1
					f := pageparser.FormatFromFrontMatterType(it.Type)
					if f == metadecoders.YAML {
						offset -= 1
					}
					pos.LineNumber += offset

					_ = fe.UpdatePosition(pos)
					_ = fe.SetFilename("") // It will be set later.

					return fe
				}

				return err
			}
			next := iter.Peek()
			if !next.IsDone() {
				p.posMainContent = next.Pos()
			}
			// Done.
			break Loop
		case it.IsEOF():
			break Loop
		case it.IsError():
			return p.failMap(source, it.Err, it)
		default:

		}
	}

	return nil
}

func (p *PageSource) failMap(source []byte, err error, i pageparser.Item) error {
	var fe herrors.FileError
	if errors.As(err, &fe) {
		return fe
	}

	pos := posFromInput("", source, i.Pos())

	return herrors.NewFileErrorFromPos(err, pos)
}

func posFromInput(filename string, input []byte, offset int) text.Position {
	if offset < 0 {
		return text.Position{
			Filename: filename,
		}
	}
	lf := []byte("\n")
	input = input[:offset]
	lineNumber := bytes.Count(input, lf) + 1
	endOfLastLine := bytes.LastIndex(input, lf)

	return text.Position{
		Filename:     filename,
		LineNumber:   lineNumber,
		ColumnNumber: offset - endOfLastLine,
		Offset:       offset,
	}
}
