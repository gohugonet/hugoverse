package cli

import (
	"flag"
	"github.com/mdfriday/hugoverse/internal/application"
	"github.com/mdfriday/hugoverse/internal/interfaces/static"
	"github.com/mdfriday/hugoverse/pkg/log"
)

type staticCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
}

func NewStaticCmd(parent *flag.FlagSet) (*staticCmd, error) {
	nCmd := &staticCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("static", flag.ExitOnError)
	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (oc *staticCmd) Usage() {
	oc.cmd.Usage()
}

func (oc *staticCmd) Run() error {
	l := log.NewStdLogger()

	publishDirFs, err := application.ServeGenerateStaticSite()
	if err != nil {
		l.Fatalf("failed to generate static sites: %v", err)
		return err
	}

	srv := static.NewFileServer(publishDirFs)

	if err := srv.Serve(); err != nil {
		l.Fatalf("failed to serve static sites: %v", err)
		return err
	}

	return nil
}
