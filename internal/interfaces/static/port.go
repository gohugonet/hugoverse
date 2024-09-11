package static

import (
	"fmt"
	"net"
	"strconv"
)

const (
	defaultFileServerPort = 1315
	defaultFileServerBind = "127.0.0.1"
)

type serverPortListener struct {
	p        int
	ln       net.Listener
	endpoint string
}

func (s *serverPortListener) bathPath() string {
	return fmt.Sprintf("http://%s", s.ln.Addr().String())
}

func newDefaultServerPortListener() *serverPortListener {
	ep := net.JoinHostPort(defaultFileServerBind, strconv.Itoa(defaultFileServerPort))

	l, err := net.Listen("tcp", ep)
	if err != nil {
		panic(err)
	}
	return &serverPortListener{
		p:        defaultFileServerPort,
		ln:       l,
		endpoint: ep,
	}
}
