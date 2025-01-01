
[![Tests on Linux, MacOS and Windows](https://github.com/bep/logg/workflows/Test/badge.svg)](https://github.com/bep/logg/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/logg)](https://goreportcard.com/report/github.com/bep/logg)
[![GoDoc](https://godoc.org/github.com/bep/logg?status.svg)](https://godoc.org/github.com/bep/logg)

This is a fork of the exellent [Apex Log](https://github.com/apex/log) library.

Main changes:

* Trim unneeded dependencies.
* Make `Fields` into a slice to preserve log order.
* Split the old `Interface` in two and remove all but one `Log` method (see below).
* This allows for lazy creation of messages in `Log(fmt.Stringer)` and ignoring fields added in `LevelLogger`s with levels below the `Logger`s.
* The pointer passed to `HandleLog` is not safe to use outside of the current log chain, and needs to be cloned with `Clone` first if that's needed.
* See [Benchmarks](#benchmarks) for more info.

This is probably the very fastest structured log library when logging is disabled:

<img width="492" alt="image" src="https://user-images.githubusercontent.com/394382/184383985-265a4b0e-ee36-405e-9dc1-4686d0cad57a.png">

> One can never have enough log libraries!

```go
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
```

## Benchmarks

Benchmarks below are borrowed and adapted from [Zap](https://github.com/uber-go/zap/tree/master/benchmarks).

### Logging at a disabled level without any structured context

```
name                                      time/op
DisabledWithoutFields/apex/log-10         33.9ns ± 0%
DisabledWithoutFields/bep/logg-10         0.28ns ± 0%
DisabledWithoutFields/sirupsen/logrus-10  6.54ns ± 0%
DisabledWithoutFields/rs/zerolog-10       0.31ns ± 0%

name                                      alloc/op
DisabledWithoutFields/apex/log-10           112B ± 0%
DisabledWithoutFields/bep/logg-10          0.00B
DisabledWithoutFields/sirupsen/logrus-10   16.0B ± 0%
DisabledWithoutFields/rs/zerolog-10        0.00B

name                                      allocs/op
DisabledWithoutFields/apex/log-10           1.00 ± 0%
DisabledWithoutFields/bep/logg-10           0.00
DisabledWithoutFields/sirupsen/logrus-10    1.00 ± 0%
DisabledWithoutFields/rs/zerolog-10         0.00
```


### Logging at a disabled level with some accumulated context

```
name                                           time/op
DisabledAccumulatedContext/apex/log-10         0.29ns ± 0%
DisabledAccumulatedContext/bep/logg-10         0.27ns ± 0%
DisabledAccumulatedContext/sirupsen/logrus-10  6.61ns ± 0%
DisabledAccumulatedContext/rs/zerolog-10       0.32ns ± 0%

name                                           alloc/op
DisabledAccumulatedContext/apex/log-10          0.00B
DisabledAccumulatedContext/bep/logg-10          0.00B
DisabledAccumulatedContext/sirupsen/logrus-10   16.0B ± 0%
DisabledAccumulatedContext/rs/zerolog-10        0.00B

name                                           allocs/op
DisabledAccumulatedContext/apex/log-10           0.00
DisabledAccumulatedContext/bep/logg-10           0.00
DisabledAccumulatedContext/sirupsen/logrus-10    1.00 ± 0%
DisabledAccumulatedContext/rs/zerolog-10         0.00
```

### Logging at a disabled level, adding context at each log site

```
name                                     time/op
DisabledAddingFields/apex/log-10          328ns ± 0%
DisabledAddingFields/bep/logg-10         0.38ns ± 0%
DisabledAddingFields/sirupsen/logrus-10   610ns ± 0%
DisabledAddingFields/rs/zerolog-10       10.5ns ± 0%

name                                     alloc/op
DisabledAddingFields/apex/log-10           886B ± 0%
DisabledAddingFields/bep/logg-10          0.00B
DisabledAddingFields/sirupsen/logrus-10  1.52kB ± 0%
DisabledAddingFields/rs/zerolog-10        24.0B ± 0%

name                                     allocs/op
DisabledAddingFields/apex/log-10           10.0 ± 0%
DisabledAddingFields/bep/logg-10           0.00
DisabledAddingFields/sirupsen/logrus-10    12.0 ± 0%
DisabledAddingFields/rs/zerolog-10         1.00 ± 0%
```

### Logging without any structured context

```
name                                    time/op
WithoutFields/apex/log-10                964ns ± 0%
WithoutFields/bep/logg-10                100ns ± 0%
WithoutFields/go-kit/kit/log-10          232ns ± 0%
WithoutFields/inconshreveable/log15-10  2.13µs ± 0%
WithoutFields/sirupsen/logrus-10         866ns ± 0%
WithoutFields/stdlib.Println-10         7.08ns ± 0%
WithoutFields/stdlib.Printf-10          56.4ns ± 0%
WithoutFields/rs/zerolog-10             30.9ns ± 0%
WithoutFields/rs/zerolog.Formatting-10  1.33µs ± 0%
WithoutFields/rs/zerolog.Check-10       32.1ns ± 0%

name                                    alloc/op
WithoutFields/apex/log-10                 352B ± 0%
WithoutFields/bep/logg-10                56.0B ± 0%
WithoutFields/go-kit/kit/log-10           520B ± 0%
WithoutFields/inconshreveable/log15-10  1.43kB ± 0%
WithoutFields/sirupsen/logrus-10        1.14kB ± 0%
WithoutFields/stdlib.Println-10          16.0B ± 0%
WithoutFields/stdlib.Printf-10            136B ± 0%
WithoutFields/rs/zerolog-10              0.00B
WithoutFields/rs/zerolog.Formatting-10  1.92kB ± 0%
WithoutFields/rs/zerolog.Check-10        0.00B

name                                    allocs/op
WithoutFields/apex/log-10                 6.00 ± 0%
WithoutFields/bep/logg-10                 2.00 ± 0%
WithoutFields/go-kit/kit/log-10           9.00 ± 0%
WithoutFields/inconshreveable/log15-10    20.0 ± 0%
WithoutFields/sirupsen/logrus-10          23.0 ± 0%
WithoutFields/stdlib.Println-10           1.00 ± 0%
WithoutFields/stdlib.Printf-10            6.00 ± 0%
WithoutFields/rs/zerolog-10               0.00
WithoutFields/rs/zerolog.Formatting-10    58.0 ± 0%
WithoutFields/rs/zerolog.Check-10         0.00
```


### Logging with some accumulated context

```
name                                         time/op
AccumulatedContext/apex/log-10               12.7µs ± 0%
AccumulatedContext/bep/logg-10               1.52µs ± 0%
AccumulatedContext/go-kit/kit/log-10         2.52µs ± 0%
AccumulatedContext/inconshreveable/log15-10  9.36µs ± 0%
AccumulatedContext/sirupsen/logrus-10        3.41µs ± 0%
AccumulatedContext/rs/zerolog-10             37.9ns ± 0%
AccumulatedContext/rs/zerolog.Check-10       34.0ns ± 0%
AccumulatedContext/rs/zerolog.Formatting-10  1.36µs ± 0%

name                                         alloc/op
AccumulatedContext/apex/log-10               3.30kB ± 0%
AccumulatedContext/bep/logg-10               1.16kB ± 0%
AccumulatedContext/go-kit/kit/log-10         3.67kB ± 0%
AccumulatedContext/inconshreveable/log15-10  3.31kB ± 0%
AccumulatedContext/sirupsen/logrus-10        4.73kB ± 0%
AccumulatedContext/rs/zerolog-10              0.00B
AccumulatedContext/rs/zerolog.Check-10        0.00B
AccumulatedContext/rs/zerolog.Formatting-10  1.92kB ± 0%

name                                         allocs/op
AccumulatedContext/apex/log-10                 53.0 ± 0%
AccumulatedContext/bep/logg-10                 25.0 ± 0%
AccumulatedContext/go-kit/kit/log-10           56.0 ± 0%
AccumulatedContext/inconshreveable/log15-10    70.0 ± 0%
AccumulatedContext/sirupsen/logrus-10          68.0 ± 0%
AccumulatedContext/rs/zerolog-10               0.00
AccumulatedContext/rs/zerolog.Check-10         0.00
AccumulatedContext/rs/zerolog.Formatting-10    58.0 ± 0%
```


## Logging with additional context at each log site

```
name                                   time/op
AddingFields/apex/log-10               13.2µs ± 0%
AddingFields/bep/logg-10               1.79µs ± 0%
AddingFields/go-kit/kit/log-10         2.23µs ± 0%
AddingFields/inconshreveable/log15-10  14.3µs ± 0%
AddingFields/sirupsen/logrus-10        4.46µs ± 0%
AddingFields/rs/zerolog-10              398ns ± 0%
AddingFields/rs/zerolog.Check-10        389ns ± 0%

name                                   alloc/op
AddingFields/apex/log-10               4.19kB ± 0%
AddingFields/bep/logg-10               2.02kB ± 0%
AddingFields/go-kit/kit/log-10         3.31kB ± 0%
AddingFields/inconshreveable/log15-10  6.68kB ± 0%
AddingFields/sirupsen/logrus-10        6.27kB ± 0%
AddingFields/rs/zerolog-10              24.0B ± 0%
AddingFields/rs/zerolog.Check-10        24.0B ± 0%

name                                   allocs/op
AddingFields/apex/log-10                 63.0 ± 0%
AddingFields/bep/logg-10                 34.0 ± 0%
AddingFields/go-kit/kit/log-10           57.0 ± 0%
AddingFields/inconshreveable/log15-10    74.0 ± 0%
AddingFields/sirupsen/logrus-10          79.0 ± 0%
AddingFields/rs/zerolog-10               1.00 ± 0%
AddingFields/rs/zerolog.Check-10         1.00 ± 0%
```
