package logg

import (
	"fmt"
	"time"
)

// Logger is the main interface for the logger.
type Logger interface {
	// WithLevel returns a new entry with `level` set.
	WithLevel(Level) *Entry
}

// LevelLogger is the logger at a given level.
type LevelLogger interface {
	// Log logs a message at the given level using the string from calling s.String().
	// Note that s.String() will not be called if the level is not enabled.
	Log(s fmt.Stringer)

	// Logf logs a message at the given level using the format and args from calling fmt.Sprintf().
	// Note that fmt.Sprintf() will not be called if the level is not enabled.
	Logf(format string, a ...any)

	// WithLevel returns a new entry with `level` set.
	WithLevel(Level) *Entry

	// WithFields returns a new entry with the`fields` in fields set.
	// This is a noop if LevelLogger's level is less than Logger's.
	WithFields(fields Fielder) *Entry

	// WithLevel returns a new entry with the field f set with value v
	// This is a noop if LevelLogger's level is less than Logger's.
	WithField(f string, v any) *Entry

	// WithDuration returns a new entry with the "duration" field set
	// to the given duration in milliseconds.
	// This is a noop if LevelLogger's level is less than Logger's.
	WithDuration(time.Duration) *Entry

	// WithError returns a new entry with the "error" set to `err`.
	// This is a noop if err is nil or  LevelLogger's level is less than Logger's.
	WithError(error) *Entry
}
