package cli

import (
	"flag"
	"github.com/mdfriday/hugoverse/internal/application"
	"github.com/mdfriday/hugoverse/pkg/log"
)

type buildCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
}

func NewBuildCmd(parent *flag.FlagSet) (*buildCmd, error) {
	nCmd := &buildCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("build", flag.ExitOnError)
	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (oc *buildCmd) Usage() {
	oc.cmd.Usage()
}

func (oc *buildCmd) Run() error {
	l := log.NewStdLogger()

	if err := application.GenerateStaticSite(); err != nil {
		l.Fatalf("failed to generate static sites: %v", err)
		return err
	}

	return nil
}
