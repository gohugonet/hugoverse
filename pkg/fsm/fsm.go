package fsm

func New(initState State, initData Data) FSM {
	return &fsm{
		state:    initState,
		data:     initData,
		handlers: map[State]StateHandler{},
	}
}

type fsm struct {
	state    State
	data     Data
	handlers map[State]StateHandler
}

func (f *fsm) State() State {
	return f.state
}

func (f *fsm) Add(state State, handler StateHandler) {
	if _, ok := f.handlers[state]; ok {
		panic("state handler exist already")
	}
	f.handlers[state] = handler
}

func (f *fsm) Process(message any) error {
	h, ok := f.handlers[f.state]
	if !ok {
		panic("state handler not exist")
	}
	s, d := h(&event{
		t:       Action,
		data:    f.data,
		message: message,
	})
	f.state = s
	f.data = d
	return f.data.Error()
}
