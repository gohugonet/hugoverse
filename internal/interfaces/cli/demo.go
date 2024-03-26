package cli

import (
	"flag"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
)

type demoCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
}

func NewDemoCmd(parent *flag.FlagSet) (*demoCmd, error) {
	nCmd := &demoCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("demo", flag.ExitOnError)
	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (oc *demoCmd) Usage() {
	oc.cmd.Usage()
}

func (oc *demoCmd) Run() error {
	dir, err := application.NewDemo()
	if err != nil {
		return err
	}
	fmt.Printf("demo dir: %s\n", dir)

	return nil
}
