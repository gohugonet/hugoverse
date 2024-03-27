package executor

import (
	"github.com/gohugonet/hugoverse/pkg/template/parser"
)

func evalActionNode(c context, n parser.Node) (context, error) {
	return context{
		state: stateAction,
		rcv:   c.rcv,
		w:     c.w,
		last:  missingVal,
	}, nil
}
