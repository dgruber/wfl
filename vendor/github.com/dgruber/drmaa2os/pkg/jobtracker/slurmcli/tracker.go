package slurmcli

import (
	"errors"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/helper"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
)

// Tracker implements the JobTracker interface by calling
// the slurm command line.
type Tracker struct {
	processTracker *simpletracker.JobTracker
	sessionName    string
	slurm          *Slurm
}

// New creates a new Tracker for slurm jobs using the
// given command line apps.
func New(jobsession string, slurmCLI *Slurm) (*Tracker, error) {
	if err := CheckCLI(slurmCLI); err != nil {
		return nil, err
	}
	// TODO at the moment a slurm account represents a job session
	// must exist before...
	return &Tracker{
		processTracker: simpletracker.New(jobsession),
		sessionName:    jobsession,
		slurm:          slurmCLI,
	}, nil
}

// ListJobs shows all running slurm jobs.
func (t *Tracker) ListJobs() ([]string, error) {
	return t.slurm.ListJobs(t.sessionName, "all")
}

// AddJob submits a new slurm job.
func (t *Tracker) AddJob(jt drmaa2interface.JobTemplate) (string, error) {
	return t.slurm.SubmitJob(t.sessionName, jt)
}

// AddArrayJob creates a slurm job array.
func (t *Tracker) AddArrayJob(jt drmaa2interface.JobTemplate, begin, end, step, maxParallel int) (string, error) {
	return "", nil
}

// ListArrayJobs shows all slums jobs which belong to a certain job array.
func (t *Tracker) ListArrayJobs(arrayjobid string) ([]string, error) {
	return nil, nil
}

// JobState returns the state of the slum job.
func (t *Tracker) JobState(jobid string) drmaa2interface.JobState {
	return t.slurm.State(t.sessionName, jobid)
}

// JobInfo returns detailed information about the job.
func (t *Tracker) JobInfo(jobid string) (drmaa2interface.JobInfo, error) {
	return drmaa2interface.JobInfo{}, nil
}

// JobControl suspends, resumes, or stops a slurm job.
func (t *Tracker) JobControl(jobid, state string) error {
	switch state {
	case "suspend":
		return t.slurm.Suspend(t.sessionName, jobid)
	case "resume":
		return t.slurm.Resume(t.sessionName, jobid)
	case "hold":
		return errors.New("Unsupported Operation")
	case "release":
		return errors.New("Unsupported Operation")
	case "terminate":
		return t.slurm.Terminate(t.sessionName, jobid)
	}
	return errors.New("undefined state")
}

// Wait blocks until either one of the given states is reached or when
// the timeout occurs.
func (t *Tracker) Wait(jobid string, timeout time.Duration, states ...drmaa2interface.JobState) error {
	return helper.WaitForState(t, jobid, timeout, states...)
}

// ListJobCategories returns nothing.
func (t *Tracker) ListJobCategories() ([]string, error) {
	return []string{}, nil
}

// DeleteJob removes the job from the internal storage. It errors
// when the job is not yet in any end state.
func (t *Tracker) DeleteJob(jobid string) error {
	return nil
}
