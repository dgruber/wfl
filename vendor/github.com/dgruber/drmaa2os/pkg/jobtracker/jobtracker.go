package jobtracker

import (
	"github.com/dgruber/drmaa2interface"
	"time"
)

type JobTracker interface {
	ListJobs() ([]string, error)
	ListArrayJobs(string) ([]string, error)
	AddJob(jt drmaa2interface.JobTemplate) (string, error)
	AddArrayJob(jt drmaa2interface.JobTemplate, begin int, end int, step int, maxParallel int) (string, error)
	JobState(jobid string) drmaa2interface.JobState
	JobInfo(jobid string) (drmaa2interface.JobInfo, error)
	JobControl(jobid, state string) error
	Wait(jobid string, timeout time.Duration, state ...drmaa2interface.JobState) error
	DeleteJob(jobid string) error
	ListJobCategories() ([]string, error)
}
