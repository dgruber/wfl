package wfl

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"time"
)

type element struct {
	job              drmaa2interface.Job
	template         drmaa2interface.JobTemplate
	terminated       bool
	submitError      error
	terminationError error
	jobinfo          drmaa2interface.JobInfo
	jobinfoError     error
	retry            int
}

type Job struct {
	wfl       *Workflow
	joblist   []*element // predecessors
	lastError error
}

// NewJob creates the initial empty job with the given workflow.
func NewJob(wfl *Workflow) *Job {
	return &Job{
		wfl:     wfl,
		joblist: make([]*element, 0),
	}
}

func EmptyJob() *Job {
	return &Job{}
}

func (j *Job) lastJob() *element {
	if len(j.joblist) == 0 {
		return nil
	}
	return j.joblist[len(j.joblist)-1]
}

func (j *Job) jobCheck() (drmaa2interface.Job, error) {
	if element := j.lastJob(); element == nil {
		return nil, errors.New("job element not available")
	} else if element.job == nil {
		return nil, errors.New("job not available")
	} else {
		return element.job, nil
	}
}

// Job Properties

// Template returns the JobTemplate of the previous job submission.
func (j *Job) Template() *drmaa2interface.JobTemplate {
	j.lastError = nil
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		if template, errTmp := job.GetJobTemplate(); errTmp != nil {
			j.lastError = errTmp
		} else {
			return &template
		}
	}
	return nil
}

// State returns the current state of the job previously submitted.
func (j *Job) State() drmaa2interface.JobState {
	job, err := j.jobCheck()
	if err != nil {
		j.lastError = err
		return drmaa2interface.Undetermined
	}
	return job.GetState()
}

// ------------
// NON-Blocking
// ------------

func (j *Job) checkCtx() error {
	if j.wfl == nil {
		return errors.New("no workflow defined")
	}
	if j.wfl.ctx == nil {
		return errors.New("no context defined")
	}
	return nil
}

// Run submits a job which executes the given command and args. Needs
// to be available on the execution backend.
func (j *Job) Run(cmd string, args ...string) *Job {
	jt := drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args}
	return j.RunT(jt)
}

// Run submits a job given specified with the JobTemplate.
func (j *Job) RunT(jt drmaa2interface.JobTemplate) *Job {
	if err := j.checkCtx(); err != nil {
		j.lastError = err
		return j
	}

	if jt.JobCategory == "" {
		// TODO RunT should not know about Docker
		jt.JobCategory = j.wfl.ctx.defaultDockerImage
	}
	job, err := j.wfl.js.RunJob(jt)
	j.lastError = err
	j.joblist = append(j.joblist, &element{job: job, submitError: err, template: jt})
	return j
}

func (j *Job) Do(f func(job drmaa2interface.Job)) *Job {
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		f(job)
	}
	return j
}

// Suspend stops a job from execution.
func (j *Job) Suspend() *Job {
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		j.lastError = job.Suspend()
	}
	return j
}

// Resume continues a suspended job to continue execution.
func (j *Job) Resume() *Job {
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		j.lastError = job.Resume()
	}
	return j
}

// Kill stops the job from execution.
func (j *Job) Kill() *Job {
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		j.lastError = job.Terminate()
	}
	return j
}

// Returns the error if occured during last job operation.
func (j *Job) LastError() error {
	return j.lastError
}

func (j *Job) Resubmit(r int) *Job {
	for i := 0; i < r; i++ {
		if e := j.lastJob(); e != nil {
			job, err := j.wfl.js.RunJob(e.template)
			j.lastError = err
			if err == nil {
				j.joblist = append(j.joblist, &element{job: job, submitError: err, template: e.template})
			}
		} else {
			j.lastError = errors.New("job not available")
			break
		}
	}
	return j
}

// OneFailed returns true when one job in the whole chain failed.
func (j *Job) OneFailed() bool {
	for _, element := range j.joblist {
		if element.job.GetState() == drmaa2interface.Failed {
			return true
		}
	}
	return false
}

// Blocking

// After blocks the given duration and continues by returning the same job.
func (j *Job) After(d time.Duration) *Job {
	<-time.After(d)
	return j
}

// Wait until the most recently job was finished.
func (j *Job) Wait() *Job {
	j.lastError = nil
	if element := j.lastJob(); element != nil && element.job != nil {
		element.terminationError = element.job.WaitTerminated(drmaa2interface.InfiniteTime)
		element.terminated = true
		element.jobinfo, element.jobinfoError = element.job.GetJobInfo()
	} else {
		j.lastError = errors.New("job not available")
	}
	return j
}

// Retry waits until the last job in chain (not for the previous ones) is finished.
// When it failed it resubmits it and waits again for a successful end.
func (j *Job) Retry(r int) *Job {
	for ; r > 0; r-- {
		if j.Wait().Success() {
			return j
		}
		j.Resubmit(1)
	}
	return j
}

// Synchronize with all jobs in the chain. All needs to be terminated until
// the call returns.
func (j *Job) Synchronize() *Job {
	for _, element := range j.joblist {
		element.job.WaitTerminated(drmaa2interface.InfiniteTime)
	}
	return j
}

// Failed returns true in case the current job stated equals drmaa2interface.Failed
func (j *Job) Failed() bool {
	if j.State() == drmaa2interface.Failed {
		return true
	}
	return false
}

// Failed returns true in case the current job stated equals drmaa2interface.Done
// and the job exit status is 0.
func (j *Job) Success() bool {
	if j.State() == drmaa2interface.Done {
		if j.ExitStatus() == 0 {
			return true
		}
	}
	return false
}

// ExitStatus waits until the previously submitted job is finished and
// returns the exit status of the job. In case of an internal error it
// returns -1.
func (j *Job) ExitStatus() int {
	j.Wait()
	if element := j.lastJob(); element != nil {
		return element.jobinfo.ExitStatus
	}
	return -1
}

// Then waits until the previous job is terminated and executes the
// given function by providing the DRMAA2 job interface.
func (j *Job) Then(f func(job drmaa2interface.Job)) *Job {
	j.lastError = nil
	if element := j.lastJob(); element != nil && element.job != nil {
		element.terminationError = element.job.WaitTerminated(drmaa2interface.InfiniteTime)
		element.terminated = true
		element.jobinfo, element.jobinfoError = element.job.GetJobInfo()
		f(element.job)
	} else {
		j.lastError = errors.New("job not available")
	}
	return j
}

// ThenRun waits until the previous job is terminated and executes then
// the given command as new job.
func (j *Job) ThenRun(cmd string, args ...string) *Job {
	return j.Wait().Run(cmd, args...)
}

// ThenRun waits until the previous job is terminated and executes then
// a new job based on the given JobTemplate.
func (j *Job) ThenRunT(jt drmaa2interface.JobTemplate) *Job {
	return j.Wait().RunT(jt)
}

func waitForJobEndAndState(j *Job) drmaa2interface.JobState {
	job, err := j.jobCheck()
	if err != nil {
		return drmaa2interface.Undetermined
	} else {
		lastError := job.WaitTerminated(drmaa2interface.InfiniteTime)
		if lastError != nil {
			return drmaa2interface.Undetermined
		}
		return job.GetState()
	}
	return drmaa2interface.Undetermined
}

// OnSuccess executes the given function after the previously submitted
// job finished in the drmaa2interface.Done state.
func (j *Job) OnSuccess(f func(job drmaa2interface.Job)) *Job {
	if waitForJobEndAndState(j) == drmaa2interface.Done {
		j.Then(f)
	}
	return j
}

// OnSuccessRun submits a job when the previous job ended in the
// job state drmaa2interface.Done.
func (j *Job) OnSuccessRun(cmd string, args ...string) *Job {
	if waitForJobEndAndState(j) == drmaa2interface.Done {
		j.Run(cmd, args...)
	}
	return j
}

// OnFailure executes the given function when the previous job in the list failed.
// Fails mean the job was started successfully by the system but then existed with
// an exit code != 0.
//
// When running the job resulted in an error (i.e. the job run function errored),
// then the function is not executed.
func (j *Job) OnFailure(f func(job drmaa2interface.Job)) *Job {
	if waitForJobEndAndState(j) != drmaa2interface.Done {
		j.Then(f)
	}
	return j
}

// OnFailureRun submits a job when the previous job ended in a state
// different than drmaa2interface.Done.
func (j *Job) OnFailureRun(cmd string, args ...string) *Job {
	if waitForJobEndAndState(j) != drmaa2interface.Done {
		j.Run(cmd, args...)
	}
	return j
}

// OnError executes the given function if the last Job operation resulted
// in an error (like a job submission failure).
func (j *Job) OnError(f func(err error)) *Job {
	if j.lastError != nil {
		f(j.lastError)
	}
	return j
}
