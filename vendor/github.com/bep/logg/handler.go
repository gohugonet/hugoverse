package logg

// Handler is used to handle log events, outputting them to
// stdio or sending them to remote services. See the "handlers"
// directory for implementations.
//
// It is left up to Handlers to implement thread-safety.
type Handler interface {
	// HandleLog is invoked for each log event.
	// Note that if the Entry is going to be used after the call to HandleLog
	// in the handler chain returns, it must be cloned with Clone(). See
	// the memory.Handler implementation for an example.
	//
	// The Entry can be modified if needed, e.g. when passed down via
	// a multi.Handler (e.g. to sanitize the data).
	HandleLog(e *Entry) error
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as
// log handlers. If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(*Entry) error

// HandleLog calls f(e).
func (f HandlerFunc) HandleLog(e *Entry) error {
	return f(e)
}
