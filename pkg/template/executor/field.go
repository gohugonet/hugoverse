package executor

import (
	"github.com/gohugonet/hugoverse/pkg/template/parser"
)

func evalFieldNode(c context, n parser.Node) (context, error) {

	ptr := c.rcv.data()
	field := n.String()
	method := ptr.MethodByName(field[1:]) // 0 is .

	v := evalCall(method, c.last)

	return context{
		state: stateCommand,
		rcv:   c.rcv,
		w:     c.w,
		last:  v,
	}, nil
}
