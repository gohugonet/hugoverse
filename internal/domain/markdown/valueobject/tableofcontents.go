// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/collections"
	"html/template"
	"sort"
	"strings"
)

// Empty is an empty ToC.
var Empty = &Fragments{
	Headings:    Headings{},
	HeadingsMap: map[string]*TocHeading{},
}

// TocBuilder is used to build the ToC data structure.
type TocBuilder struct {
	toc *Fragments
}

// AddAt adds the heading to the ToC.
func (b *TocBuilder) AddAt(h *TocHeading, row, level int) {
	if b.toc == nil {
		b.toc = &Fragments{}
	}
	b.toc.addAt(h, row, level)
}

// Build returns the ToC.
func (b TocBuilder) Build() *Fragments {
	if b.toc == nil {
		return Empty
	}
	b.toc.HeadingsMap = make(map[string]*TocHeading)
	b.toc.walk(func(h *TocHeading) {
		if h.ID != "" {
			b.toc.HeadingsMap[h.ID] = h
			b.toc.Identifiers = append(b.toc.Identifiers, h.ID)
		}
	})
	sort.Strings(b.toc.Identifiers)
	return b.toc
}

// Headings holds the top level headings.
type Headings []*TocHeading

// FilterBy returns a new Headings slice with all headings that matches the given predicate.
// For internal use only.
func (h Headings) FilterBy(fn func(*TocHeading) bool) Headings {
	var out Headings

	for _, h := range h {
		h.walk(func(h *TocHeading) {
			if fn(h) {
				out = append(out, h)
			}
		})
	}
	return out
}

// TocHeading holds the data about a heading and its children.
type TocHeading struct {
	ID    string
	Level int
	Title string

	Headings Headings
}

// IsZero is true when no ID or Text is set.
func (h TocHeading) IsZero() bool {
	return h.ID == "" && h.Title == ""
}

func (h *TocHeading) walk(fn func(*TocHeading)) {
	fn(h)
	for _, h := range h.Headings {
		h.walk(fn)
	}
}

// Fragments holds the table of contents for a page.
type Fragments struct {
	// Headings holds the top level headings.
	Headings Headings

	// Identifiers holds all the identifiers in the ToC as a sorted slice.
	// Note that collections.SortedStringSlice has both a Contains and Count method
	// that can be used to identify missing and duplicate IDs.
	Identifiers collections.SortedStringSlice

	// HeadingsMap holds all the headings in the ToC as a map.
	// Note that with duplicate IDs, the last one will win.
	HeadingsMap map[string]*TocHeading
}

// addAt adds the heading into the given location.
func (toc *Fragments) addAt(h *TocHeading, row, level int) {
	for i := len(toc.Headings); i <= row; i++ {
		toc.Headings = append(toc.Headings, &TocHeading{})
	}

	if level == 0 {
		toc.Headings[row] = h
		return
	}

	heading := toc.Headings[row]

	for i := 1; i < level; i++ {
		if len(heading.Headings) == 0 {
			heading.Headings = append(heading.Headings, &TocHeading{})
		}
		heading = heading.Headings[len(heading.Headings)-1]
	}
	heading.Headings = append(heading.Headings, h)
}

// ToHTML renders the ToC as HTML.
func (toc *Fragments) ToHTML(startLevel, stopLevel int, ordered bool) template.HTML {
	if toc == nil {
		return ""
	}
	b := &tocBuilder{
		s:          strings.Builder{},
		h:          toc.Headings,
		startLevel: startLevel,
		stopLevel:  stopLevel,
		ordered:    ordered,
	}
	b.Build()
	return template.HTML(b.s.String())
}

func (toc Fragments) walk(fn func(*TocHeading)) {
	for _, h := range toc.Headings {
		h.walk(fn)
	}
}

type tocBuilder struct {
	s strings.Builder
	h Headings

	startLevel int
	stopLevel  int
	ordered    bool
}

func (b *tocBuilder) Build() {
	b.writeNav(b.h)
}

func (b *tocBuilder) writeNav(h Headings) {
	b.s.WriteString("<nav id=\"TableOfContents\">")
	b.writeHeadings(1, 0, b.h)
	b.s.WriteString("</nav>")
}

func (b *tocBuilder) writeHeadings(level, indent int, h Headings) {
	if level < b.startLevel {
		for _, h := range h {
			b.writeHeadings(level+1, indent, h.Headings)
		}
		return
	}

	if b.stopLevel != -1 && level > b.stopLevel {
		return
	}

	hasChildren := len(h) > 0

	if hasChildren {
		b.s.WriteString("\n")
		b.indent(indent + 1)
		if b.ordered {
			b.s.WriteString("<ol>\n")
		} else {
			b.s.WriteString("<ul>\n")
		}
	}

	for _, h := range h {
		b.writeHeading(level+1, indent+2, h)
	}

	if hasChildren {
		b.indent(indent + 1)
		if b.ordered {
			b.s.WriteString("</ol>")
		} else {
			b.s.WriteString("</ul>")
		}
		b.s.WriteString("\n")
		b.indent(indent)
	}
}

func (b *tocBuilder) writeHeading(level, indent int, h *TocHeading) {
	b.indent(indent)
	b.s.WriteString("<li>")
	if !h.IsZero() {
		b.s.WriteString("<a href=\"#" + h.ID + "\">" + h.Title + "</a>")
	}
	b.writeHeadings(level, indent, h.Headings)
	b.s.WriteString("</li>\n")
}

func (b *tocBuilder) indent(n int) {
	for i := 0; i < n; i++ {
		b.s.WriteString("  ")
	}
}

// DefaultTocConfig is the default ToC configuration.
var DefaultTocConfig = TocConfig{
	StartLevel: 2,
	EndLevel:   3,
	Ordered:    false,
}

type TocConfig struct {
	// Heading start level to include in the table of contents, starting
	// at h1 (inclusive).
	// <docsmeta>{ "identifiers": ["h1"] }</docsmeta>
	StartLevel int

	// Heading end level, inclusive, to include in the table of contents.
	// Default is 3, a value of -1 will include everything.
	EndLevel int

	// Whether to produce a ordered list or not.
	Ordered bool
}
