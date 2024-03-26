package cli

import (
	"flag"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/interfaces/api"
	"github.com/gohugonet/hugoverse/pkg/log"
	"strconv"
)

type serverCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
	port   *string
	env    *string
	https  *bool
}

func NewServeCmd(parent *flag.FlagSet) (*serverCmd, error) {
	nCmd := &serverCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("normal", flag.ExitOnError)
	nCmd.port = nCmd.cmd.String("port", "1314",
		fmt.Sprintln("[optional] server listening port, default is `1314`"))
	nCmd.env = nCmd.cmd.String("env", "dev",
		fmt.Sprintln("[optional, dev|prod] development environment, default is `dev`"))
	nCmd.https = nCmd.cmd.Bool("https", false,
		fmt.Sprintln("[optional] enable https, default is `false`"))

	err := nCmd.cmd.Parse(parent.Args()[1:])
	if err != nil {
		return nil, err
	}

	return nCmd, nil
}

func (c *serverCmd) Usage() {
	c.cmd.Usage()
}

func (c *serverCmd) Run() error {
	l := log.NewStdLogger()

	env := api.DEV
	if *c.env == "prod" {
		env = api.PROD
	}
	s, err := api.NewServer(setupLogger(l), setupPort(*c.port))
	if err != nil {
		l.Fatalf("Error creating server: %v", err)
	}
	defer s.Close()

	l.Fatalf("Error listening on :%v: %v", *c.port, s.ListenAndServe(env, *c.https))

	return nil
}

func setupLogger(l log.Logger) func(s *api.Server) error {
	return func(s *api.Server) error {
		s.Log = l

		return nil
	}
}

func setupPort(port string) func(s *api.Server) error {
	return func(s *api.Server) error {
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		s.HttpPort = p

		return nil
	}
}
