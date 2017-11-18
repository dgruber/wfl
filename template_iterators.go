package wfl

import (
	"github.com/dgruber/drmaa2interface"
	"strconv"
)

// NewEnvSequenceIterator returns an iterator which increments the
// environment variable env each time when called.
func NewEnvSequenceIterator(env string, start, incr int) Iterator {
	sequence := start
	return func(t drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
		if t.JobEnvironment == nil {
			t.JobEnvironment = make(map[string]string, 1)
		}
		t.JobEnvironment[env] = strconv.Itoa(sequence)
		sequence += incr
		return t
	}
}
