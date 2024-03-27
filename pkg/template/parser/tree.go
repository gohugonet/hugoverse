package parser

type tree struct {
	*treeNode
}

func (t *tree) String() string {
	cs := t.Children()
	s := ""
	for _, n := range cs {
		s += n.(Node).String()
	}
	return s
}

func (t *tree) Type() NodeType {
	return RootNode
}

func newTree() *tree {
	return &tree{treeNode: &treeNode{nodes: []TreeNode{}}}
}

type treeNode struct {
	nodes []TreeNode
	baseNode
}

func (n *treeNode) AppendChild(node TreeNode) {
	n.nodes = append(n.nodes, node)
}

func (n *treeNode) Children() []TreeNode {
	return n.nodes
}

func (t *tree) Walk(walker Walker) {
	walkNode(t, walker)
}

func walkNode(n Node, walker Walker) WalkStatus {
	status := walker(n, WalkIn)
	if status != WalkStop {
		cs := n.Children()
		for _, c := range cs {
			if s := walkNode(c.(Node), walker); s == WalkStop {
				return WalkStop
			}
		}
	}
	return walker(n, WalkOut)
}
