package valueobject

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"strconv"
	"sync"
)

type ShortcodeParser struct {
	shortcodes []*Shortcode

	// All the shortcode names in this set.
	nameSet   map[string]bool
	nameSetMu sync.RWMutex

	source []byte

	pid     uint64
	ordinal int
	level   int

	tmplSvc contenthub.Template
}

func NewShortcodeParser(source []byte, pid uint64, tmplSvc contenthub.Template) *ShortcodeParser {
	return &ShortcodeParser{
		nameSet: make(map[string]bool),
		source:  source,

		pid:     pid,
		ordinal: 0,
		level:   0,

		tmplSvc: tmplSvc,
	}
}

func (s *ShortcodeParser) ParseItem(it pageparser.Item, pt *pageparser.Iterator) (*Shortcode, error) {
	currShortcode, err := s.extractShortcode(s.ordinal, 0, pt)
	if err != nil {
		return nil, err
	}

	currShortcode.Pos = it.Pos()
	currShortcode.Length = pt.Current().Pos() - it.Pos()

	if currShortcode.Name != "" {
		s.addName(currShortcode.Name)
	}

	if currShortcode.Params == nil {
		var s []string
		currShortcode.Params = s
	}

	currShortcode.Placeholder = createShortcodePlaceholder("s", s.pid, s.ordinal)
	s.ordinal++
	s.shortcodes = append(s.shortcodes, currShortcode)

	return currShortcode, nil
}

// Note - this value must not contain any markup syntax
const shortcodePlaceholderPrefix = "HAHAHUGOSHORTCODE"

func createShortcodePlaceholder(sid string, id uint64, ordinal int) string {
	return shortcodePlaceholderPrefix + strconv.FormatUint(id, 10) + sid + strconv.Itoa(ordinal) + "HBHB"
}

// pageTokens state:
// - before: positioned just before the shortcode start
// - after: shortcode(s) consumed (plural when they are nested)
func (s *ShortcodeParser) extractShortcode(ordinal, level int, pt *pageparser.Iterator) (*Shortcode, error) {
	if s == nil {
		panic("handler nil")
	}
	sc := &Shortcode{Ordinal: ordinal}

	// Back up one to identify any indentation.
	if pt.Pos() > 0 {
		pt.Backup()
		item := pt.Next()
		if item.IsIndentation() {
			sc.Indentation = item.ValStr(s.source)
		}
	}

	cnt := 0
	nestedOrdinal := 0
	nextLevel := level + 1
	closed := false
	const errorPrefix = "failed to extract shortcode"

Loop:
	for {
		currItem := pt.Next()
		switch {
		case currItem.IsLeftShortcodeDelim():
			next := pt.Peek()
			if next.IsRightShortcodeDelim() {
				// no name: {{< >}} or {{% %}}
				return sc, errors.New("shortcode has no name")
			}
			if next.IsShortcodeClose() {
				continue
			}

			if cnt > 0 {
				// nested shortcode; append it to inner content
				pt.Backup()
				nested, err := s.extractShortcode(nestedOrdinal, nextLevel, pt)
				nestedOrdinal++
				if nested != nil && nested.Name != "" {
					s.addName(nested.Name)
				}

				if err == nil {
					sc.Inner = append(sc.Inner, nested)
				} else {
					return sc, err
				}

			} else {
				sc.DoMarkup = currItem.IsShortcodeMarkupDelimiter()
			}

			cnt++

		case currItem.IsRightShortcodeDelim():
			// we trust the template on this:
			// if there's no inner, we're done
			if !sc.IsInline {
				if !sc.Info.ParseInfo().Inner() {
					return sc, nil
				}
			}

		case currItem.IsShortcodeClose():
			closed = true
			next := pt.Peek()
			if !sc.IsInline {
				if !sc.NeedsInner() {
					if next.IsError() {
						// return that error, more specific
						continue
					}
					return nil, fmt.Errorf("%s: shortcode %q does not evaluate .Inner or .InnerDeindent, yet a closing tag was provided", errorPrefix, next.ValStr(s.source))
				}
			}
			if next.IsRightShortcodeDelim() {
				// self-closing
				pt.Consume(1)
			} else {
				sc.IsClosing = true
				pt.Consume(2)
			}

			return sc, nil
		case currItem.IsText():
			sc.Inner = append(sc.Inner, currItem.ValStr(s.source))
		case currItem.IsShortcodeName():

			sc.Name = currItem.ValStr(s.source)

			// Used to check if the template expects inner content.
			templs := s.tmplSvc.LookupVariants(sc.Name)
			if templs == nil {
				return nil, fmt.Errorf("%s: template for shortcode %q not found", errorPrefix, sc.Name)
			}

			sc.Info = templs[0].(template.Info)
			sc.Templs = templs
		case currItem.IsInlineShortcodeName():
			sc.Name = currItem.ValStr(s.source)
			sc.IsInline = true
		case currItem.IsShortcodeParam():
			if !pt.IsValueNext() {
				continue
			} else if pt.Peek().IsShortcodeParamVal() {
				// named params
				if sc.Params == nil {
					params := make(map[string]any)
					params[currItem.ValStr(s.source)] = pt.Next().ValTyped(s.source)
					sc.Params = params
				} else {
					if params, ok := sc.Params.(map[string]any); ok {
						params[currItem.ValStr(s.source)] = pt.Next().ValTyped(s.source)
					} else {
						return sc, fmt.Errorf("%s: invalid state: invalid param type %T for shortcode %q, expected a map", errorPrefix, params, sc.Name)
					}
				}
			} else {
				// positional params
				if sc.Params == nil {
					var params []any
					params = append(params, currItem.ValTyped(s.source))
					sc.Params = params
				} else {
					if params, ok := sc.Params.([]any); ok {
						params = append(params, currItem.ValTyped(s.source))
						sc.Params = params
					} else {
						return sc, fmt.Errorf("%s: invalid state: invalid param type %T for shortcode %q, expected a slice", errorPrefix, params, sc.Name)
					}
				}
			}
		case currItem.IsDone():
			if !currItem.IsError() {
				if !closed && sc.NeedsInner() {
					return sc, fmt.Errorf("%s: shortcode %q must be closed or self-closed", errorPrefix, sc.Name)
				}
			}
			// handled by caller
			pt.Backup()
			break Loop

		}
	}
	return sc, nil
}

func (s *ShortcodeParser) addName(name string) {
	s.nameSetMu.Lock()
	defer s.nameSetMu.Unlock()
	s.nameSet[name] = true
}
