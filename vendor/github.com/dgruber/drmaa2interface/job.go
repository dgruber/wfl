package drmaa2interface

import (
	"time"
)

// Job defines all methods which needs to be implemented for
// a DRMAA job type.
type Job interface {
	GetID() string
	GetSessionName() string
	GetJobTemplate() (JobTemplate, error)
	GetState() JobState
	GetJobInfo() (JobInfo, error)
	Suspend() error
	Resume() error
	Hold() error
	Release() error
	Terminate() error
	WaitStarted(time.Duration) error
	WaitTerminated(time.Duration) error
	Reap() error
}

// JobState represents the state of a job.
type JobState int

//go:generate stringer -type=JobState
const (
	Unset JobState = iota
	Undetermined
	Queued
	QueuedHeld
	Running
	Suspended
	Requeued
	RequeuedHeld
	Done
	Failed
)
