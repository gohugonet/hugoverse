package fsm

type Data interface {
	Error() error
	Raw() any
}

type State string
type StateHandler func(event Event) (State, Data)

type TypeEvent int

type Event interface {
	Type() TypeEvent
	Message() any
	Data() Data
}

type FSM interface {
	Add(state State, handler StateHandler)
	Process(message any) error
	State() State
}
