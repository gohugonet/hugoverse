package cli

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/pkg/log"
	"os"
)

type buildCmd struct {
	parent       *flag.FlagSet
	cmd          *flag.FlagSet
	hugoProjPath *string
}

func NewBuildCmd(parent *flag.FlagSet) (*buildCmd, error) {
	nCmd := &buildCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("build", flag.ExitOnError)
	nCmd.hugoProjPath = nCmd.cmd.String("p", "", fmt.Sprintf(
		"[required] target hugo project pathspec \n(e.g. %s)", "pathspec/to/your/hugo/project"))
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

	if *oc.hugoProjPath == "" {
		oc.cmd.Usage()
		return errors.New("please specify a target hugo project path spec")
	}

	_, err := os.Stat(*oc.hugoProjPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", *oc.hugoProjPath)
	}

	if err != nil {
		return err
	}

	if err = application.GenerateStaticSite(*oc.hugoProjPath); err != nil {
		l.Fatalf("failed to generate static sites: %v", err)
		return err
	}

	return nil
}
