package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

type Job struct {
	id       string
	session  string
	template drmaa2interface.JobTemplate
	tracker  jobtracker.JobTracker // reference to external job tracker
}

func NewJob(id, session string, jt drmaa2interface.JobTemplate, tracker jobtracker.JobTracker) *Job {
	return &Job{
		id:       id,
		session:  session,
		template: jt,
		tracker:  tracker,
	}
}

func (j *Job) GetID() string {
	return j.id
}

func (j *Job) GetSessionName() string {
	return j.session
}

func (j *Job) GetJobTemplate() (drmaa2interface.JobTemplate, error) {
	return j.template, nil
}

func (j *Job) GetJobInfo() (drmaa2interface.JobInfo, error) {
	return j.tracker.JobInfo(j.id)
}

func (j *Job) GetState() drmaa2interface.JobState {
	return j.tracker.JobState(j.id)
}

func (j *Job) Suspend() error {
	return j.tracker.JobControl(j.id, "suspend")
}

func (j *Job) Resume() error {
	return j.tracker.JobControl(j.id, "resume")
}

func (j *Job) Hold() error {
	return j.tracker.JobControl(j.id, "hold")
}

func (j *Job) Release() error {
	return j.tracker.JobControl(j.id, "release")
}

func (j *Job) Terminate() error {
	return j.tracker.JobControl(j.id, "terminate")
}

func (j *Job) WaitStarted(timeout time.Duration) error {
	return j.tracker.Wait(j.id, timeout, drmaa2interface.Running, drmaa2interface.Failed, drmaa2interface.Done)
}

func (j *Job) WaitTerminated(timeout time.Duration) error {
	return j.tracker.Wait(j.id, timeout, drmaa2interface.Done, drmaa2interface.Failed)
}

func (j *Job) Reap() error {
	state := j.tracker.JobState(j.id)
	if state != drmaa2interface.Done && state != drmaa2interface.Failed {
		return ErrorInvalidState
	}
	return j.tracker.DeleteJob(j.id)
}
