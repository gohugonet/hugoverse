package fsm

const (
	Notification TypeEvent = iota
	Action
)

type event struct {
	t       TypeEvent
	data    Data
	message any
}

func (e *event) Type() TypeEvent {
	return e.t
}
func (e *event) Data() Data {
	return e.data
}
func (e *event) Message() any {
	return e.message
}
func (e *event) Error() error {
	return e.data.Error()
}
