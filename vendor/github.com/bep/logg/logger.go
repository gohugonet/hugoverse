package logg

import (
	"fmt"
	stdlog "log"
	"time"

	"github.com/bep/clocks"
)

// assert interface compliance.
var _ Logger = (*logger)(nil)

// String implements fmt.Stringer and can be used directly in
// the log methods.
type String string

// StringFunc is a function that returns a string.
// It also implements the fmt.Stringer interface and
// can therefore be used as argument to the log methods.
type StringFunc func() string

func (f StringFunc) String() string {
	return f()
}

func (s String) String() string {
	return string(s)
}

// Fielder is an interface for providing fields to custom types.
type Fielder interface {
	Fields() Fields
}

func NewFieldsFunc(fn func() Fields) FieldsFunc {
	return FieldsFunc(fn)
}

type FieldsFunc func() Fields

func (f FieldsFunc) Fields() Fields {
	return f()
}

// Field holds a named value.
type Field struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// Fields represents a slice of entry level data used for structured logging.
type Fields []Field

// Fields implements Fielder.
func (f Fields) Fields() Fields {
	return f
}

// Options is the set of options used to configure a logger.
type Options struct {
	// Level is the minimum level to log at.
	// If not set, defaults to InfoLevel.
	Level Level

	// Handler is the log handler to use.
	Handler Handler

	// Clock is the clock to use for timestamps.
	// If not set, the system clock is used.
	Clock Clock
}

// New returns a new logger.
func New(cfg Options) Logger {
	if cfg.Handler == nil {
		panic("handler cannot be nil")
	}

	if cfg.Level <= 0 || cfg.Level > LevelError {
		panic("log level is out of range")
	}

	if cfg.Clock == nil {
		cfg.Clock = clocks.System()
	}

	if cfg.Level == 0 {
		cfg.Level = LevelInfo
	}

	return &logger{
		Handler: cfg.Handler,
		Level:   cfg.Level,
		Clock:   cfg.Clock,
	}
}

// logger represents a logger with configurable Level and Handler.
type logger struct {
	Handler Handler
	Level   Level
	Clock   Clock
}

// Clock provides the current time.
type Clock interface {
	Now() time.Time
}

// WithLevel returns a new entry with `level` set.
func (l *logger) WithLevel(level Level) *Entry {
	return NewEntry(l).WithLevel(level)
}

// WithFields returns a new entry with `fields` set.
func (l *logger) WithFields(fields Fielder) *Entry {
	return NewEntry(l).WithFields(fields.Fields())
}

// WithField returns a new entry with the `key` and `value` set.
//
// Note that the `key` should not have spaces in it - use camel
// case or underscores
func (l *logger) WithField(key string, value any) *Entry {
	return NewEntry(l).WithField(key, value)
}

// WithDuration returns a new entry with the "duration" field set
// to the given duration in milliseconds.
func (l *logger) WithDuration(d time.Duration) *Entry {
	return NewEntry(l).WithDuration(d)
}

// WithError returns a new entry with the "error" set to `err`.
func (l *logger) WithError(err error) *Entry {
	return NewEntry(l).WithError(err)
}

// ErrStopLogEntry is a sentinel error that can be returned from a
// handler to stop the entry from being passed to the next handler.
var ErrStopLogEntry = fmt.Errorf("stop log entry")

// log the message, invoking the handler.
func (l *logger) log(e *Entry, s fmt.Stringer) {
	if e.Level < l.Level {
		return
	}

	finalized := objectPools.GetEntry()
	defer objectPools.PutEntry(finalized)
	e.finalize(finalized, s.String())

	if err := l.Handler.HandleLog(finalized); err != nil {
		if err != ErrStopLogEntry {
			stdlog.Printf("error logging: %s", err)
		}
	}
}
