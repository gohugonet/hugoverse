package fmt

import (
	_fmt "fmt"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"sort"

	"github.com/bep/logg"
	"github.com/spf13/cast"
)

// New returns a new instance of the fmt-namespaced template functions.
func New() *Namespace {
	ns := &Namespace{
		logger: loggers.NewDefault(),
	}

	return ns
}

// Namespace provides template functions for the "fmt" namespace.
type Namespace struct {
	logger loggers.Logger
}

// Print returns a string representation of args.
func (ns *Namespace) Print(args ...any) string {
	return _fmt.Sprint(args...)
}

// Printf returns string representation of args formatted with the layout in format.
func (ns *Namespace) Printf(format string, args ...any) string {
	return _fmt.Sprintf(format, args...)
}

// Println returns string representation of args  ending with a newline.
func (ns *Namespace) Println(args ...any) string {
	return _fmt.Sprintln(args...)
}

// Errorf formats args according to a format specifier and logs an ERROR.
// It returns an empty string.
func (ns *Namespace) Errorf(format string, args ...any) string {
	ns.logger.Errorf(format, args...)
	return ""
}

// Erroridf formats args according to a format specifier and logs an ERROR and
// an information text that the error with the given id can be suppressed in config.
// It returns an empty string.
func (ns *Namespace) Erroridf(id, format string, args ...any) string {
	ns.logger.Erroridf(id, format, args...)
	return ""
}

// Warnf formats args according to a format specifier and logs a WARNING.
// It returns an empty string.
func (ns *Namespace) Warnf(format string, args ...any) string {
	ns.logger.Warnf(format, args...)
	return ""
}

// Warnidf formats args according to a format specifier and logs an WARNING and
// an information text that the warning with the given id can be suppressed in config.
// It returns an empty string.
func (ns *Namespace) Warnidf(id, format string, args ...any) string {
	ns.logger.Warnidf(id, format, args...)
	return ""
}

// Warnmf is experimental and subject to change at any time.
func (ns *Namespace) Warnmf(m any, format string, args ...any) string {
	return ns.logmf(ns.logger.Warn(), m, format, args...)
}

// Errormf is experimental and subject to change at any time.
func (ns *Namespace) Errormf(m any, format string, args ...any) string {
	return ns.logmf(ns.logger.Error(), m, format, args...)
}

func (ns *Namespace) logmf(l logg.LevelLogger, m any, format string, args ...any) string {
	mm := cast.ToStringMap(m)
	fields := make(logg.Fields, len(mm))
	i := 0
	for k, v := range mm {
		fields[i] = logg.Field{Name: k, Value: v}
		i++
	}
	// Sort the fields to make the output deterministic.
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	l.WithFields(fields).Logf(format, args...)

	return ""
}
