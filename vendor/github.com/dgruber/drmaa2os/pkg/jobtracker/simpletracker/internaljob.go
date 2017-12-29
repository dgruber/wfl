package simpletracker

import (
	"github.com/dgruber/drmaa2interface"
)

type InternalJob struct {
	TaskID int
	State  drmaa2interface.JobState
	PID    int
}
