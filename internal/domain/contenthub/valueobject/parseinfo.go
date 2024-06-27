package valueobject

import (
	"bytes"
	"errors"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/text"
)

type ItemSourceHandler func(item pageparser.Item, source []byte) error
type ItemHandler func(item pageparser.Item) error
type IterHandler func(item pageparser.Item, iter *pageparser.Iterator) error

type SourceParseInfo struct {
	source []byte

	posMainContent int

	// Items from the page parser.
	// These maps directly to the source
	ItemsStep1 pageparser.Items

	FrontMatterHandler ItemSourceHandler
	SummaryHandler     IterHandler
	BytesHandler       ItemHandler
	ShortcodeHandler   IterHandler
}

func (s *SourceParseInfo) IsEmpty() bool {
	return len(s.ItemsStep1) == 0
}

func (s *SourceParseInfo) Handle() error {
	if s.IsEmpty() {
		return nil
	}

	iter := pageparser.NewIterator(s.ItemsStep1)

Loop:
	for {
		it := iter.Next()
		switch {
		case it.Type == pageparser.TypeIgnore:
		case it.IsFrontMatter():
			if err := s.FrontMatterHandler(it, s.source); err != nil {
				var fe herrors.FileError
				if errors.As(err, &fe) {
					pos := fe.Position()

					// Offset the starting position of front matter.
					offset := iter.LineNumber(s.source) - 1
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
				s.posMainContent = next.Pos()
			}

		case it.Type == pageparser.TypeLeadSummaryDivider:
			if err := s.SummaryHandler(it, iter); err != nil {
				return err
			}

			// Handle shortcode
		case it.IsLeftShortcodeDelim():
			// let extractShortcode handle left delim (will do so recursively)
			iter.Backup()

			if err := s.ShortcodeHandler(it, iter); err != nil {
				return s.failMap(err, it)
			}

		case it.IsEOF():
			break Loop
		case it.IsError():
			return s.failMap(it.Err, it)
		default:
			if err := s.BytesHandler(it); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SourceParseInfo) failMap(err error, i pageparser.Item) error {
	var fe herrors.FileError
	if errors.As(err, &fe) {
		return fe
	}

	pos := posFromInput("", s.source, i.Pos())

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
