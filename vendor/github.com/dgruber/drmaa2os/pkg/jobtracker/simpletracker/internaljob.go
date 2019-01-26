package simpletracker

import (
	"github.com/dgruber/drmaa2interface"
)

// InternalJob represents a process.
type InternalJob struct {
	TaskID int
	State  drmaa2interface.JobState
	PID    int
}
