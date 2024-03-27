package parser

import "github.com/gohugonet/hugoverse/pkg/lexer"

type ParseState int

const (
	done ParseState = 1 + iota
	open
)

type Parser interface {
	Parse(token lexer.Token) (Node, ParseState, error)
}

type NodeType int

const (
	RootNode NodeType = 1 + iota
	TextNode
	ActionNode
	CommandNode
	FieldNode
	IdentifierNode
)

type Nodes []Node

type Node interface {
	String() string
	Type() NodeType
	SetVal(v string)
	TreeNode
}

type TreeNode interface {
	AppendChild(node TreeNode)
	Children() []TreeNode
}

type WalkStatus int

const (
	WalkStop WalkStatus = iota + 1
	WalkContinue
)

type WalkState int

const (
	WalkIn WalkState = 1 << iota
	WalkOut
)

type Walker func(n Node, ws WalkState) WalkStatus
