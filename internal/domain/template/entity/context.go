package entity

import (
	"errors"
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/template"
	"github.com/mdfriday/hugoverse/internal/domain/template/valueobject"
	"github.com/mdfriday/hugoverse/pkg/maps"
	htmltemplate "github.com/mdfriday/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/mdfriday/hugoverse/pkg/template/texttemplate"
	"github.com/mdfriday/hugoverse/pkg/template/texttemplate/parse"
	"github.com/mitchellh/mapstructure"
)

type Context struct {
	visited          map[string]bool
	TemplateNotFound map[string]bool
	lookupFn         func(name string) *valueobject.State

	// The last error encountered.
	err error

	// Set when we're done checking for config header.
	configChecked bool

	t *valueobject.State

	// Store away the return node in partials.
	returnNode *parse.CommandNode
}

func newTemplateContext(t *valueobject.State, lookupFn func(name string) *valueobject.State) *Context {
	return &Context{
		t:                t,
		lookupFn:         lookupFn,
		visited:          make(map[string]bool),
		TemplateNotFound: make(map[string]bool),
	}
}

func unwrap(templ template.Preparer) template.Preparer {
	if ts, ok := templ.(*valueobject.State); ok {
		return ts.Preparer
	}
	return templ
}

func getParseTree(templ template.Preparer) *parse.Tree {
	templ = unwrap(templ)
	if text, ok := templ.(*texttemplate.Template); ok {
		return text.Tree
	}
	return templ.(*htmltemplate.Template).Tree
}

func ApplyTemplateTransformers(t *valueobject.State, lookupFn func(name string) *valueobject.State) (*Context, error) {
	if t == nil {
		return nil, errors.New("expected template, but none provided")
	}

	c := newTemplateContext(t, lookupFn)
	tree := getParseTree(t.Preparer)

	_, err := c.applyTransformations(tree.Root)

	if err == nil && c.returnNode != nil {
		// This is a partial with a return statement.
		c.t.PInfo.HasReturn = true
		tree.Root = c.wrapInPartialReturnWrapper(tree.Root)
	}

	return c, err
}

// applyTransformations do 2 things:
// 1) Parses partial return statement.
// 2) Tracks template (partial) dependencies and some other info.
func (c *Context) applyTransformations(n parse.Node) (bool, error) {
	switch x := n.(type) {
	case *parse.ListNode:
		if x != nil {
			c.applyTransformationsToNodes(x.Nodes...)
		}
	case *parse.ActionNode:
		c.applyTransformationsToNodes(x.Pipe)
	case *parse.IfNode:
		c.applyTransformationsToNodes(x.Pipe, x.List, x.ElseList)
	case *parse.WithNode:
		c.applyTransformationsToNodes(x.Pipe, x.List, x.ElseList)
	case *parse.RangeNode:
		c.applyTransformationsToNodes(x.Pipe, x.List, x.ElseList)
	case *parse.TemplateNode:
		subTempl := c.getIfNotVisited(x.Name)
		if subTempl != nil {
			c.applyTransformationsToNodes(getParseTree(subTempl.Preparer).Root)
		}
	case *parse.PipeNode:
		c.collectConfig(x)
		for i, cmd := range x.Cmds {
			keep, _ := c.applyTransformations(cmd)
			if !keep {
				x.Cmds = append(x.Cmds[:i], x.Cmds[i+1:]...)
			}
		}

	case *parse.CommandNode:
		c.collectInner(x)
		keep := c.collectReturnNode(x)

		for _, elem := range x.Args {
			switch an := elem.(type) {
			case *parse.PipeNode:
				c.applyTransformations(an)
			}
		}
		return keep, c.err
	}

	return true, c.err
}

func (c *Context) applyTransformationsToNodes(nodes ...parse.Node) {
	for _, node := range nodes {
		c.applyTransformations(node)
	}
}

func (c Context) getIfNotVisited(name string) *valueobject.State {
	if c.visited[name] {
		return nil
	}
	c.visited[name] = true
	templ := c.lookupFn(name)
	if templ == nil {
		// This may be a inline template defined outside of this file
		// and not yet parsed. Unusual, but it happens.
		// Store the name to try again later.
		c.TemplateNotFound[name] = true
	}

	return templ
}

// collectConfig collects and parses any leading template config variable declaration.
// This will be the first PipeNode in the template, and will be a variable declaration
// on the form:
//
//	{{ $_hugo_config:= `{ "version": 1 }` }}
func (c *Context) collectConfig(n *parse.PipeNode) {
	if c.t.Typ != template.TypeShortcode {
		return
	}
	if c.configChecked {
		return
	}
	c.configChecked = true

	if len(n.Decl) != 1 || len(n.Cmds) != 1 {
		// This cannot be a config declaration
		return
	}

	v := n.Decl[0]

	if len(v.Ident) == 0 || v.Ident[0] != "$_hugo_config" {
		return
	}

	cmd := n.Cmds[0]

	if len(cmd.Args) == 0 {
		return
	}

	if s, ok := cmd.Args[0].(*parse.StringNode); ok {
		errMsg := "failed to decode $_hugo_config in template: %w"
		m, err := maps.ToStringMapE(s.Text)
		if err != nil {
			c.err = fmt.Errorf(errMsg, err)
			return
		}
		if err := mapstructure.WeakDecode(m, &c.t.PInfo.Config); err != nil {
			c.err = fmt.Errorf(errMsg, err)
		}
	}
}

// collectInner determines if the given CommandNode represents a
// shortcode call to its .Inner.
func (c *Context) collectInner(n *parse.CommandNode) {
	if c.t.Typ != template.TypeShortcode {
		return
	}
	if c.t.PInfo.IsInner || len(n.Args) == 0 {
		return
	}

	for _, arg := range n.Args {
		var idents []string
		switch nt := arg.(type) {
		case *parse.FieldNode:
			idents = nt.Ident
		case *parse.VariableNode:
			idents = nt.Ident
		}

		if c.hasIdent(idents, "Inner") || c.hasIdent(idents, "InnerDeindent") {
			c.t.PInfo.IsInner = true
			break
		}
	}
}

func (c *Context) hasIdent(idents []string, ident string) bool {
	for _, id := range idents {
		if id == ident {
			return true
		}
	}
	return false
}

func (c *Context) collectReturnNode(n *parse.CommandNode) bool {
	if c.t.Typ != template.TypePartial || c.returnNode != nil {
		return true
	}

	if len(n.Args) < 2 {
		return true
	}

	ident, ok := n.Args[0].(*parse.IdentifierNode)
	if !ok || ident.Ident != "return" {
		return true
	}

	c.returnNode = n
	// Remove the "return" identifiers
	c.returnNode.Args = c.returnNode.Args[1:]

	return false
}

const (
	// We parse this template and modify the nodes in order to assign
	// the return value of a partial to a contextWrapper via Set. We use
	// "range" over a one-element slice so we can shift dot to the
	// partial's argument, Arg, while allowing Arg to be falsy.
	partialReturnWrapperTempl = `{{ $_hugo_dot := $ }}{{ $ := .Arg }}{{ range (slice .Arg) }}{{ $_hugo_dot.Set ("PLACEHOLDER") }}{{ end }}`
)

var partialReturnWrapper *parse.ListNode

func init() {
	templ, err := texttemplate.New("").Parse(partialReturnWrapperTempl)
	if err != nil {
		panic(err)
	}
	partialReturnWrapper = templ.Tree.Root
}

// wrapInPartialReturnWrapper copies and modifies the parsed nodes of a
// predefined partial return wrapper to insert those of a user-defined partial.
func (c *Context) wrapInPartialReturnWrapper(n *parse.ListNode) *parse.ListNode {
	wrapper := partialReturnWrapper.CopyList()
	rangeNode := wrapper.Nodes[2].(*parse.RangeNode)
	retn := rangeNode.List.Nodes[0]
	setCmd := retn.(*parse.ActionNode).Pipe.Cmds[0]
	setPipe := setCmd.Args[1].(*parse.PipeNode)
	// Replace PLACEHOLDER with the real return value.
	// Note that this is a PipeNode, so it will be wrapped in parens.
	setPipe.Cmds = []*parse.CommandNode{c.returnNode}
	rangeNode.List.Nodes = append(n.Nodes, retn)

	return wrapper
}
