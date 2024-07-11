package valueobject

import (
	"bytes"
	"errors"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/text"
)

type ItemHandler func(item pageparser.Item) error
type IterHandler func(item pageparser.Item, iter *pageparser.Iterator) error

type SourceHandlers interface {
	FrontMatterHandler() ItemHandler
	SummaryHandler() IterHandler
	BytesHandler() ItemHandler
	ShortcodeHandler() IterHandler
}

type SourceParseInfo struct {
	Source []byte

	// TODO: should belongs to the content
	posMainContent int

	// Items from the page parser.
	// These maps directly to the source
	ItemsStep1 pageparser.Items

	Handlers SourceHandlers
}

func NewSourceParseInfo(source []byte, handlers SourceHandlers) (*SourceParseInfo, error) {
	if fmh := handlers.FrontMatterHandler(); fmh == nil {
		return nil, errors.New("no front matter handler")
	}
	if sh := handlers.SummaryHandler(); sh == nil {
		return nil, errors.New("no summary handler")
	}
	if sch := handlers.ShortcodeHandler(); sch == nil {
		return nil, errors.New("no shortcode handler")
	}
	if bh := handlers.BytesHandler(); bh == nil {
		return nil, errors.New("no bytes handler")
	}

	return &SourceParseInfo{
		Source:         source,
		Handlers:       handlers,
		posMainContent: -1,
	}, nil
}

func (s *SourceParseInfo) IsEmpty() bool {
	return len(s.ItemsStep1) == 0
}

func (s *SourceParseInfo) Parse() error {
	items, err := pageparser.ParseBytes(
		s.Source,
		pageparser.Config{},
	)
	if err != nil {
		return err
	}

	s.ItemsStep1 = items
	return nil
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
			if err := s.Handlers.FrontMatterHandler()(it); err != nil {
				var fe herrors.FileError
				if errors.As(err, &fe) {
					pos := fe.Position()

					// Offset the starting position of front matter.
					offset := iter.LineNumber(s.Source) - 1
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
			if err := s.Handlers.SummaryHandler()(it, iter); err != nil {
				return err
			}

			// Handle shortcode
		case it.IsLeftShortcodeDelim():
			// let extractShortcode handle left delim (will do so recursively)
			iter.Backup()

			if err := s.Handlers.ShortcodeHandler()(it, iter); err != nil {
				return s.failMap(err, it)
			}

		case it.IsEOF():
			break Loop
		case it.IsError():
			return s.failMap(it.Err, it)
		default:
			if err := s.Handlers.BytesHandler()(it); err != nil {
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

	pos := posFromInput("", s.Source, i.Pos())

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
