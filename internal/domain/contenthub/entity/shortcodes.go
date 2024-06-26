package entity

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"sync"
)

type shortcodeContentStore interface {
	AddShortcode(s *valueobject.Shortcode)
}

type Shortcodes struct {
	shortcodes []*valueobject.Shortcode

	// All the shortcode names in this set.
	nameSet   map[string]bool
	nameSetMu sync.RWMutex

	source []byte

	ordinal int
	level   int

	contentStore shortcodeContentStore
}

func (s *Shortcodes) shortcodeHandler(pt *pageparser.Iterator) error {
	currShortcode, err := s.extractShortcode(s.ordinal, 0, pt)
	if err != nil {
		return fail(err, it)
	}

	currShortcode.pos = it.Pos()
	currShortcode.length = iter.Current().Pos() - it.Pos()
	if currShortcode.placeholder == "" {
		currShortcode.placeholder = createShortcodePlaceholder("s", rn.pid, currShortcode.ordinal)
	}

	if currShortcode.name != "" {
		s.addName(currShortcode.name)
	}

	if currShortcode.params == nil {
		var s []string
		currShortcode.params = s
	}

	currShortcode.placeholder = createShortcodePlaceholder("s", rn.pid, ordinal)
	s.ordinal++
	s.shortcodes = append(s.shortcodes, currShortcode)

	s.contentStore.AddShortcode(currShortcode)

	return nil
}

// pageTokens state:
// - before: positioned just before the shortcode start
// - after: shortcode(s) consumed (plural when they are nested)
func (s *Shortcodes) extractShortcode(ordinal, level int, pt *pageparser.Iterator) (*valueobject.Shortcode, error) {
	if s == nil {
		panic("handler nil")
	}
	sc := &valueobject.Shortcode{Ordinal: ordinal}

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
				if !sc.info.ParseInfo().IsInner {
					return sc, nil
				}
			}

		case currItem.IsShortcodeClose():
			closed = true
			next := pt.Peek()
			if !sc.IsInline {
				if !sc.needsInner() {
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
				sc.isClosing = true
				pt.Consume(2)
			}

			return sc, nil
		case currItem.IsText():
			sc.inner = append(sc.inner, currItem.ValStr(source))
		case currItem.IsShortcodeName():

			sc.name = currItem.ValStr(source)

			// Used to check if the template expects inner content.
			templs := s.s.Tmpl().LookupVariants(sc.name)
			if templs == nil {
				return nil, fmt.Errorf("%s: template for shortcode %q not found", errorPrefix, sc.name)
			}

			sc.info = templs[0].(tpl.Info)
			sc.templs = templs
		case currItem.IsInlineShortcodeName():
			sc.name = currItem.ValStr(source)
			sc.isInline = true
		case currItem.IsShortcodeParam():
			if !pt.IsValueNext() {
				continue
			} else if pt.Peek().IsShortcodeParamVal() {
				// named params
				if sc.params == nil {
					params := make(map[string]any)
					params[currItem.ValStr(source)] = pt.Next().ValTyped(source)
					sc.params = params
				} else {
					if params, ok := sc.params.(map[string]any); ok {
						params[currItem.ValStr(source)] = pt.Next().ValTyped(source)
					} else {
						return sc, fmt.Errorf("%s: invalid state: invalid param type %T for shortcode %q, expected a map", errorPrefix, params, sc.name)
					}
				}
			} else {
				// positional params
				if sc.params == nil {
					var params []any
					params = append(params, currItem.ValTyped(source))
					sc.params = params
				} else {
					if params, ok := sc.params.([]any); ok {
						params = append(params, currItem.ValTyped(source))
						sc.params = params
					} else {
						return sc, fmt.Errorf("%s: invalid state: invalid param type %T for shortcode %q, expected a slice", errorPrefix, params, sc.name)
					}
				}
			}
		case currItem.IsDone():
			if !currItem.IsError() {
				if !closed && sc.needsInner() {
					return sc, fmt.Errorf("%s: shortcode %q must be closed or self-closed", errorPrefix, sc.name)
				}
			}
			// handled by caller
			pt.Backup()
			break Loop

		}
	}
	return sc, nil
}

func (s *Shortcodes) addName(name string) {
	s.nameSetMu.Lock()
	defer s.nameSetMu.Unlock()
	s.nameSet[name] = true
}
