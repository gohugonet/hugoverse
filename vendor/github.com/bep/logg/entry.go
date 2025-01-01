package logg

import (
	"fmt"
	"strings"
	"time"
)

// assert interface compliance.
var (
	_ LevelLogger = (*Entry)(nil)
)

// Entry represents a single log entry at a given log level.
type Entry struct {
	logger *logger

	Level     Level     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Fields    Fields    `json:"fields,omitempty"`
	Message   string    `json:"message"`

	fieldsAddedCounter int
}

// NewEntry returns a new entry for `log`.
func NewEntry(log *logger) *Entry {
	return &Entry{
		logger: log,
	}
}

func (e Entry) WithLevel(level Level) *Entry {
	e.Level = level
	return &e
}

func (e *Entry) WithFields(fielder Fielder) *Entry {
	if e.isLevelDisabled() {
		return e
	}
	x := *e
	fields := fielder.Fields()
	x.fieldsAddedCounter += len(fields)
	x.Fields = append(x.Fields, fields...)
	if x.fieldsAddedCounter > 100 {
		// This operation will eventually also be performed on the final entry,
		// do it here to avoid the slice to grow indefinitely.
		x.mergeFields()
		x.fieldsAddedCounter = 0
	}
	return &x
}

func (e *Entry) WithField(key string, value any) *Entry {
	if e.isLevelDisabled() {
		return e
	}
	return e.WithFields(Fields{{key, value}})
}

func (e *Entry) WithDuration(d time.Duration) *Entry {
	if e.isLevelDisabled() {
		return e
	}
	return e.WithField("duration", d.Milliseconds())
}

// WithError returns a new entry with the "error" set to `err`.
//
// The given error may implement .Fielder, if it does the method
// will add all its `.Fields()` into the returned entry.
func (e *Entry) WithError(err error) *Entry {
	if err == nil || e.isLevelDisabled() {
		return e
	}

	ctx := e.WithField("error", err.Error())

	if s, ok := err.(stackTracer); ok {
		frame := s.StackTrace()[0]

		name := fmt.Sprintf("%n", frame)
		file := fmt.Sprintf("%+s", frame)
		line := fmt.Sprintf("%d", frame)

		parts := strings.Split(file, "\n\t")
		if len(parts) > 1 {
			file = parts[1]
		}

		ctx = ctx.WithField("source", fmt.Sprintf("%s: %s:%s", name, file, line))
	}

	if f, ok := err.(Fielder); ok {
		ctx = ctx.WithFields(f)
	}

	return ctx
}

func (e *Entry) isLevelDisabled() bool {
	return e.Level < e.logger.Level
}

// Log a message at the given level.
func (e *Entry) Log(s fmt.Stringer) {
	e.logger.log(e, s)
}

// Log a message at the given level.
func (e *Entry) Logf(format string, a ...any) {
	e.logger.log(e, StringFunc(func() string {
		return fmt.Sprintf(format, a...)
	}))

}

// Clone returns a new Entry with the same fields.
func (e *Entry) Clone() *Entry {
	x := *e
	x.Fields = make(Fields, len(e.Fields))
	copy(x.Fields, e.Fields)
	return &x
}

func (e *Entry) reset() {
	e.logger = nil
	e.Level = 0
	e.Fields = e.Fields[:0]
	e.Message = ""
	e.Timestamp = time.Time{}
}

// Remove any early entries with the same name.
func (e *Entry) mergeFields() {
	n := 0
	for i, f := range e.Fields {
		keep := true
		for j := i + 1; j < len(e.Fields); j++ {
			if e.Fields[j].Name == f.Name {
				keep = false
				break
			}
		}
		if keep {
			e.Fields[n] = f
			n++
		}
	}
	e.Fields = e.Fields[:n]
}

// finalize populates dst with Level and  Fields merged from e and Message and Timestamp set.
func (e *Entry) finalize(dst *Entry, msg string) {
	dst.Message = msg
	dst.Timestamp = e.logger.Clock.Now()
	dst.Level = e.Level
	if cap(dst.Fields) < len(e.Fields) {
		dst.Fields = make(Fields, len(e.Fields))
	} else {
		dst.Fields = dst.Fields[:len(e.Fields)]
	}
	copy(dst.Fields, e.Fields)
	dst.mergeFields()
}
