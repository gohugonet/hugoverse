package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/text"
	"html/template"
	"reflect"
	"strings"
	"sync"
)

// ShortcodeWithPage is the "." context in a shortcode template.
type ShortcodeWithPage struct {
	Params        any
	Inner         template.HTML
	Page          contenthub.Page
	Parent        *ShortcodeWithPage
	Name          string
	IsNamedParams bool

	// Zero-based ordinal in relation to its parent. If the parent is the page itself,
	// this ordinal will represent the position of this shortcode in the page content.
	Ordinal int

	// Indentation before the opening shortcode in the source.
	indentation string

	innerDeindentInit sync.Once
	innerDeindent     template.HTML

	// pos is the position in bytes in the source file. Used for error logging.
	posInit   sync.Once
	posOffset int
	pos       text.Position

	scratch *maps.Scratch
}

// InnerDeindent returns the (potentially de-indented) inner content of the shortcode.
func (scp *ShortcodeWithPage) InnerDeindent() template.HTML {
	if scp.indentation == "" {
		return scp.Inner
	}
	scp.innerDeindentInit.Do(func() {
		b := bp.GetBuffer()
		text.VisitLinesAfter(string(scp.Inner), func(s string) {
			if strings.HasPrefix(s, scp.indentation) {
				b.WriteString(strings.TrimPrefix(s, scp.indentation))
			} else {
				b.WriteString(s)
			}
		})
		scp.innerDeindent = template.HTML(b.String())
		bp.PutBuffer(b)
	})

	return scp.innerDeindent
}

// Position returns this shortcode's detailed position. Note that this information
// may be expensive to calculate, so only use this in error situations.
func (scp *ShortcodeWithPage) Position() text.Position {
	scp.posInit.Do(func() {
		if p, ok := scp.Page.(contenthub.PageContext); ok {
			scp.pos = p.PosOffset(scp.posOffset)
		}
	})
	return scp.pos
}

// Scratch returns a scratch-pad scoped for this shortcode. This can be used
// as a temporary storage for variables, counters etc.
func (scp *ShortcodeWithPage) Scratch() *maps.Scratch {
	if scp.scratch == nil {
		scp.scratch = maps.NewScratch()
	}
	return scp.scratch
}

// Get is a convenience method to look up shortcode parameters by its key.
func (scp *ShortcodeWithPage) Get(key any) any {
	if scp.Params == nil {
		return nil
	}
	if reflect.ValueOf(scp.Params).Len() == 0 {
		return nil
	}

	var x reflect.Value

	switch key.(type) {
	case int64, int32, int16, int8, int:
		if reflect.TypeOf(scp.Params).Kind() == reflect.Map {
			// We treat this as a non error, so people can do similar to
			// {{ $myParam := .Get "myParam" | default .Get 0 }}
			// Without having to do additional checks.
			return nil
		} else if reflect.TypeOf(scp.Params).Kind() == reflect.Slice {
			idx := int(reflect.ValueOf(key).Int())
			ln := reflect.ValueOf(scp.Params).Len()
			if idx > ln-1 {
				return ""
			}
			x = reflect.ValueOf(scp.Params).Index(idx)
		}
	case string:
		if reflect.TypeOf(scp.Params).Kind() == reflect.Map {
			x = reflect.ValueOf(scp.Params).MapIndex(reflect.ValueOf(key))
			if !x.IsValid() {
				return ""
			}
		} else if reflect.TypeOf(scp.Params).Kind() == reflect.Slice {
			// We treat this as a non error, so people can do similar to
			// {{ $myParam := .Get "myParam" | default .Get 0 }}
			// Without having to do additional checks.
			return nil
		}
	}

	return x.Interface()
}

func (scp *ShortcodeWithPage) UnwrapPage() contenthub.Page {
	return scp.Page
}
