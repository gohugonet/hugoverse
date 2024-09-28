package cli

import (
	"flag"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/pkg/log"
)

type loadCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
}

func NewLoadCmd(parent *flag.FlagSet) (*loadCmd, error) {
	nCmd := &loadCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("build", flag.ExitOnError)
	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (oc *loadCmd) Usage() {
	oc.cmd.Usage()
}

func (oc *loadCmd) Run() error {
	l := log.NewStdLogger()

	if err := application.LoadHugoProject(); err != nil {
		l.Fatalf("failed to generate static sites: %v", err)
		return err
	}

	return nil
}
