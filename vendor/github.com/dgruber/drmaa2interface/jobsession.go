package drmaa2interface

import (
	"time"
)

// JobSession contains all functions for managing jobs within
// one names job session. Multiple concurrent job sessions can
// be used at one point in time. A job session is a logical
// concept for separating workflows.
type JobSession interface {
	Close() error
	GetContact() (string, error)
	GetSessionName() (string, error)
	GetJobCategories() ([]string, error)
	GetJobs(ji JobInfo) ([]Job, error)
	GetJobArray(id string) (ArrayJob, error)
	RunJob(jt JobTemplate) (Job, error)
	RunBulkJobs(jt JobTemplate, begin int, end int, step int, maxParallel int) (ArrayJob, error)
	WaitAnyStarted(jobs []Job, timeout time.Duration) (Job, error)
	WaitAnyTerminated(jobs []Job, timeout time.Duration) (Job, error)
}
