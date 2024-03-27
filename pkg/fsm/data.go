package fsm

type BaseData struct {
	Err     error
	RawData string
}

func (d *BaseData) Error() error {
	return d.Err
}
func (d *BaseData) Raw() any {
	return d.RawData
}
