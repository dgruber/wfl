package wfl

import (
	"github.com/dgruber/drmaa2interface"
	"strconv"
	"time"
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

// NewTimeIterator returns a template iterator which return a job
// template every d time.
func NewTimeIterator(d time.Duration) Iterator {
	ch := time.NewTicker(d).C
	iteration := 0
	return func(t drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
		<-ch
		if t.JobEnvironment == nil {
			t.JobEnvironment = make(map[string]string, 2)
		}
		iteration++
		t.JobEnvironment["wfl_iteration"] = strconv.Itoa(iteration)
		t.JobEnvironment["wfl_time"] = time.Now().String()
		return t
	}
}
