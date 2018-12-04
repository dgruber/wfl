package singularity

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/helper"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
)

// Tracker tracks singularity container.
type Tracker struct {
	processTracker *simpletracker.JobTracker
}

// New creates a new Tracker for Singularity containers.
func New(jobsession string) (*Tracker, error) {
	cmd := exec.Command("singularity")
	err := cmd.Run()
	if err != nil || !cmd.ProcessState.Success() {
		return nil, fmt.Errorf("singularity command is not found")
	}
	return &Tracker{
		processTracker: simpletracker.New(jobsession),
	}, nil
}

// ListJobs shows all Singularity containers running.
func (dt *Tracker) ListJobs() ([]string, error) {
	return dt.processTracker.ListJobs()
}

// AddJob creates a new Singularity container.
func (dt *Tracker) AddJob(jt drmaa2interface.JobTemplate) (string, error) {
	if jt.JobCategory == "" {
		return "", fmt.Errorf("Singularity container image not specified")
	}
	return dt.processTracker.AddJob(createProcessJobTemplate(jt))
}

// AddArrayJob creates (end - begin)/step Singularity containers.
// TODO: maxParallel is not evaluated
func (dt *Tracker) AddArrayJob(jt drmaa2interface.JobTemplate, begin int, end int, step int, maxParallel int) (string, error) {
	return helper.AddArrayJobAsSingleJobs(jt, dt, begin, end, step)
}

// ListArrayJobs shows all containers which belong to a certain job array.
func (dt *Tracker) ListArrayJobs(ID string) ([]string, error) {
	return helper.ArrayJobID2GUIDs(ID)
}

// JobState returns the state of the Singularity container.
func (dt *Tracker) JobState(jobid string) drmaa2interface.JobState {
	return dt.processTracker.JobState(jobid)
}

// JobInfo returns detailed information about the job.
func (dt *Tracker) JobInfo(jobid string) (drmaa2interface.JobInfo, error) {
	return dt.processTracker.JobInfo(jobid)
}

// JobControl suspends, resumes, or stops a Singularity container.
func (dt *Tracker) JobControl(jobid, state string) error {
	return dt.processTracker.JobControl(jobid, state)
}

// Wait blocks until either one of the given states is reached or when the timeout occurs.
func (dt *Tracker) Wait(jobid string, timeout time.Duration, state ...drmaa2interface.JobState) error {
	return helper.WaitForState(dt.processTracker, jobid, timeout, state...)
}

// ListJobCategories returns nothing.
func (dt *Tracker) ListJobCategories() ([]string, error) {
	return []string{}, nil
}

// DeleteJob TODO as it is a TODO in simpletracker. Removes the job from the internal
// DB when it is finsihed.
func (dt *Tracker) DeleteJob(jobid string) error {
	return dt.processTracker.DeleteJob(jobid)
}
