package clocks

import (
	"time"
)

// TimeCupFinalNorway1976 is the start time in UTC for the final match of the 1976 Norwegian Football Cup.
// This is typically used in tests where you need a historic time with a special meaning.
var TimeCupFinalNorway1976 = time.Date(1976, time.October, 24, 12, 15, 2, 127686412, time.UTC)

// Clock provides the sub set of methods in time.Time that this package provides.
type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration

	// Offset returns the offset of this clock relative to the system clock.
	Offset() time.Duration
}

// Start creates a new Clock starting at t.
func Start(t time.Time) Clock {
	return &clock{
		offset: t.Sub(time.Now()),
	}
}

type clock struct {
	offset time.Duration
}

// Now returns the current time relative to the configured start time.
func (c *clock) Now() time.Time {
	return time.Now().Add(c.offset)
}

// Since returns the time elapsed since t.
func (c *clock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}

// Until returns the duration until t.
func (c *clock) Until(t time.Time) time.Duration {
	return t.Sub(c.Now())
}

// Offset returns the offset of this clock relative to the system clock.
// This can be used to convert to/from system time.
func (c *clock) Offset() time.Duration {
	return c.offset
}

var goClock = &systemClock{}

// System is a Clock that uses the system clock, meaning it just delegates to time.Now() etc.
func System() Clock {
	return goClock
}

type systemClock struct{}

func (c *systemClock) Now() time.Time {
	return time.Now()
}

func (c *systemClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (c *systemClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}

func (c *systemClock) Offset() time.Duration {
	return 0
}

// Fixed returns a Clock that always returns the given time.
func Fixed(t time.Time) Clock {
	return &fixedClock{t: t}
}

// fixedClock is a Clock that always returns the same time.
type fixedClock struct {
	t time.Time
}

func (c *fixedClock) Now() time.Time {
	return c.t
}

func (c *fixedClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}

func (c *fixedClock) Until(t time.Time) time.Duration {
	return t.Sub(c.Now())
}

// Offset returns the offset of this clock relative to the system clock.
func (c *fixedClock) Offset() time.Duration {
	return time.Since(c.t)
}
