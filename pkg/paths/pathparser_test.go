// Copyright 2024 The Hugo Authors. All rights reserved.
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

package paths

import (
	"github.com/mdfriday/hugoverse/pkg/paths/files"
	"testing"

	qt "github.com/frankban/quicktest"
)

var testParser = &PathParser{}

func TestParse(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		name   string
		path   string
		assert func(c *qt.C, p *Path)
	}{
		{
			"Basic Markdown file",
			"/a/b/c.md",
			func(c *qt.C, p *Path) {
				c.Assert(p.IsContent(), qt.IsTrue)
				c.Assert(p.IsLeafBundle(), qt.IsFalse)
				c.Assert(p.Name(), qt.Equals, "c.md")
				c.Assert(p.Base(), qt.Equals, "/a/b/c")
				c.Assert(p.Section(), qt.Equals, "a")
				c.Assert(p.BaseNameNoIdentifier(), qt.Equals, "c")
				c.Assert(p.Path(), qt.Equals, "/a/b/c.md")
				c.Assert(p.Dir(), qt.Equals, "/a/b")
				c.Assert(p.Container(), qt.Equals, "b")
				c.Assert(p.ContainerDir(), qt.Equals, "/a/b")
				c.Assert(p.Ext(), qt.Equals, "md")
			},
		},
		{
			"Basic text file, 1 space in dir",
			"/a b/c.txt",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a-b/c.txt")
			},
		},
		{
			"Basic text file, 2 spaces in dir",
			"/a  b/c.txt",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a--b/c.txt")
			},
		},
		{
			"Basic text file, 1 space in filename",
			"/a/b c.txt",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a/b-c.txt")
			},
		},
		{
			"Basic text file, 2 spaces in filename",
			"/a/b  c.txt",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a/b--c.txt")
			},
		},
		{
			"Basic md file, with index.md name",
			"/a/index.md",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a")
			},
		},
		{
			"Basic md file, with _index.md name",
			"/a/_index.md",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/a")
			},
		},
		{
			"Basic md file, with root _index.md name",
			"/_index.md",
			func(c *qt.C, p *Path) {
				c.Assert(p.Base(), qt.Equals, "/")
			},
		},
		{
			"Basic text file, mixed case and spaces, unnormalized",
			"/abc/Foo BAR.txt",
			func(c *qt.C, p *Path) {
				pp := p.Unnormalized()
				c.Assert(pp, qt.IsNotNil)
				c.Assert(pp.BaseNameNoIdentifier(), qt.Equals, "Foo BAR")
				c.Assert(pp.Section(), qt.Equals, "abc")
			},
		},
	}
	for _, test := range tests {
		c.Run(test.name, func(c *qt.C) {
			test.assert(c, testParser.Parse(files.ComponentFolderContent, test.path))
		})
	}

}

func TestHasExt(t *testing.T) {
	c := qt.New(t)

	c.Assert(HasExt("/a/b/c.txt"), qt.IsTrue)
	c.Assert(HasExt("/a/b.c/d.txt"), qt.IsTrue)
	c.Assert(HasExt("/a/b/c"), qt.IsFalse)
	c.Assert(HasExt("/a/b.c/d"), qt.IsFalse)
}

func BenchmarkParseIdentity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testParser.ParseIdentity(files.ComponentFolderAssets, "/a/b.css")
	}
}
