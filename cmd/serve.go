package cmd

import (
	"flag"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/interfaces/api"
	"github.com/gohugonet/hugoverse/pkg/log"
	"net/http"
)

type serverCmd struct {
	parent *flag.FlagSet
	cmd    *flag.FlagSet
	port   *string
}

func NewServeCmd(parent *flag.FlagSet) (*serverCmd, error) {
	nCmd := &serverCmd{
		parent: parent,
	}

	nCmd.cmd = flag.NewFlagSet("normal", flag.ExitOnError)
	nCmd.port = nCmd.cmd.String("port", "",
		fmt.Sprintln("[optional] server listening port"))

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
	s, err := api.NewServer(func(s *api.Server) error {
		s.Log = l

		return nil
	})
	if err != nil {
		l.Fatalf("Error creating server: %v", err)
	}

	port := *c.port
	if *c.port == "" {
		port = "1314"
	}

	l.Printf("Listening on :%v ...", port)
	l.Fatalf("Error listening on :%v: %v", port, http.ListenAndServe(":"+port, s))

	return nil
}
