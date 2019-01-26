package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

// Job represents a single computational activity that is
// executed by the DRM system. There are three relevant method sets
// for working with jobs: The JobSession interface represents all
// control and monitoring functions available for jobs. The Job
// interface represents the common control functionality for one
// existing job. Sets of jobs resulting from a bulk submission are
// controllable as a whole by the JobArray interface.
type Job struct {
	id       string
	session  string
	template drmaa2interface.JobTemplate
	tracker  jobtracker.JobTracker // reference to external job tracker
}

func newJob(id, session string, jt drmaa2interface.JobTemplate, tracker jobtracker.JobTracker) *Job {
	return &Job{
		id:       id,
		session:  session,
		template: jt,
		tracker:  tracker,
	}
}

// GetID returns the job identifier assigned by the DRM system in text form. This
// method is expected to be used as a fast alternative to the fetching of a complete
// JobInfo instance.
func (j *Job) GetID() string {
	return j.id
}

// GetSessionName reports the name of the JobSession that was used to create
// the job. If the session name cannot be determined, for example since the
// job was created outside of a DRMAA session, the attribute SHOULD be
// UNSET (i.e. equals "").
func (j *Job) GetSessionName() string {
	return j.session
}

// GetJobTemplate returns a reference to a JobTemplate instance that has
// equal values to the one that was used for the job submission creating this
// Job instance.
// For jobs created outside of a DRMAA session, implementations MUST also return a
// JobTemplate instance here, which MAY be empty or only partially filled.
func (j *Job) GetJobTemplate() (drmaa2interface.JobTemplate, error) {
	return j.template, nil
}

// GetJobInfo returns a JobInfo instance for the particular job.
func (j *Job) GetJobInfo() (drmaa2interface.JobInfo, error) {
	return j.tracker.JobInfo(j.id)
}

// GetState allows the application to get the current status of the job
// according to the DRMAA state model, together with an implementation
// specific sub state (see Section 8.1). It is intended as a fast
// alternative to the fetching of a complete JobInfo instance.
func (j *Job) GetState() drmaa2interface.JobState {
	return j.tracker.JobState(j.id)
}

// Suspend triggers a job state transition from RUNNING to SUSPENDED state.
func (j *Job) Suspend() error {
	return j.tracker.JobControl(j.id, "suspend")
}

// Resume triggers a job state transition from SUSPENDED to RUNNING state.
func (j *Job) Resume() error {
	return j.tracker.JobControl(j.id, "resume")
}

// Hold triggers a transition from QUEUED to QUEUED_HELD,
// or from REQUEUED to REQUEUED_HELD state.
func (j *Job) Hold() error {
	return j.tracker.JobControl(j.id, "hold")
}

// Release triggers a transition from QUEUED_HELD to QUEUED,
// or from REQUEUED_HELD to REQUEUED state.
func (j *Job) Release() error {
	return j.tracker.JobControl(j.id, "release")
}

// Terminate triggers a transition from any of the "Started"
// states to one of the "Terminated" states.
func (j *Job) Terminate() error {
	return j.tracker.JobControl(j.id, "terminate")
}

// WaitStarted blocks until the job entered one of the
// "Started" states.
func (j *Job) WaitStarted(timeout time.Duration) error {
	return j.tracker.Wait(j.id, timeout, drmaa2interface.Running, drmaa2interface.Failed, drmaa2interface.Done)
}

// WaitTerminated blocks until the job entered one of the "Terminated" states
func (j *Job) WaitTerminated(timeout time.Duration) error {
	return j.tracker.Wait(j.id, timeout, drmaa2interface.Done, drmaa2interface.Failed)
}

// Reap is intended to let the DRMAA implementation clean up any data
// about this job. The motivating factor are long-running applications
// maintaining large amounts of jobs as part of a monitoring session.
// Using a reaped job in any subsequent activity MUST generate an
// InvalidArgumentException for the job parameter.
// This function MUST only work for jobs in "Terminated" states, so that
// the job is promised to not change its status while being reaped.
func (j *Job) Reap() error {
	state := j.tracker.JobState(j.id)
	if state != drmaa2interface.Done && state != drmaa2interface.Failed {
		return ErrorInvalidState
	}
	return j.tracker.DeleteJob(j.id)
}
