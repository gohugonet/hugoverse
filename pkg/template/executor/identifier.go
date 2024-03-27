package executor

import (
	"github.com/gohugonet/hugoverse/pkg/template/parser"
)

func evalIdentifierNode(c context, n parser.Node) (context, error) {

	ptr := c.rcv.ptr()
	field := n.String()
	method := ptr.MethodByName(field)

	v := evalCall(method, c.last)

	return context{
		state: stateCommand,
		rcv:   c.rcv,
		w:     c.w,
		last:  v,
	}, nil
}
