[![Tests on Linux, MacOS and Windows](https://github.com/bep/clocks/workflows/Test/badge.svg)](https://github.com/bep/clocks/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/clocks)](https://goreportcard.com/report/github.com/bep/clocks)
[![GoDoc](https://godoc.org/github.com/bep/clocks?status.svg)](https://godoc.org/github.com/bep/clocks)

This package provides a _ticking clock_ that allows you to set the start time. It also provides a system clock, both implementing this interface:

```go
// Clock provides the sub set of methods in time.Time that this package provides.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration

	// Offset returns the offset of this clock relative to the system clock.
	Offset() time.Duration
}
```

Note that this only support a subset of all the methods in `time.Time` (see above) and is by design very simple. If you're looking for a more advanced time mocking library, have a look at https://github.com/benbjohnson/clock.
