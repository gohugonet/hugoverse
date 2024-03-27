package executor

import (
	"io"
	"reflect"
)

type context struct {
	state state
	rcv   *receiver
	w     io.Writer
	last  reflect.Value
}

type state uint8

const (
	stateText state = iota
	stateAction
	stateCommand
)
