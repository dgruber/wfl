package wfl

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
	"time"
)

type task struct {
	job              drmaa2interface.Job
	template         drmaa2interface.JobTemplate
	jobinfo          drmaa2interface.JobInfo
	terminated       bool
	submitError      error
	terminationError error
	jobinfoError     error
	retry            int
}

// Job defines methods for job life-cycle management. A job is
// always bound to a workflow which defines the context and
// job session (logical separation of jobs) of the underlying backend.
// The Job object allows to create an manage tasks.
type Job struct {
	wfl       *Workflow
	tasklist  []*task
	tag       string
	lastError error
}

// NewJob creates the initial empty job with the given workflow.
func NewJob(wfl *Workflow) *Job {
	return &Job{
		wfl:      wfl,
		tasklist: make([]*task, 0, 32),
	}
}

// EmptyJob creates an empty job.
func EmptyJob() *Job {
	return &Job{}
}

func (j *Job) lastJob() *task {
	if len(j.tasklist) == 0 {
		return nil
	}
	return j.tasklist[len(j.tasklist)-1]
}

func (j *Job) jobCheck() (drmaa2interface.Job, error) {
	if task := j.lastJob(); task == nil {
		return nil, errors.New("job task not available")
	} else if task.job == nil {
		return nil, errors.New("job not available")
	} else {
		return task.job, nil
	}
}

// Job Sequence Properties

// TagWith tags a job with a string for identification. Global for all tasks of the job.
func (j *Job) TagWith(tag string) *Job {
	j.tag = tag
	return j
}

// Tag returns the tag of the job.
func (j *Job) Tag() string {
	return j.tag
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

// JobID returns the job ID of the previously submitted job.
func (j *Job) JobID() string {
	job, err := j.jobCheck()
	if err != nil {
		j.lastError = err
		return ""
	}
	return job.GetID()
}

// JobInfo returns information about the last task/job. Which values
// are actually set depends on the DRMAA2 implementation of
// the backend specified in the context.
func (j *Job) JobInfo() drmaa2interface.JobInfo {
	job, err := j.jobCheck()
	if err != nil {
		j.lastError = err
		return drmaa2interface.JobInfo{}
	}
	ji, errJI := job.GetJobInfo()
	if errJI != nil {
		j.lastError = errJI
		return drmaa2interface.JobInfo{}
	}
	return ji
}

// JobInfos returns all JobInfo objects of all tasks/job run in the
// workflow. JobInfo contains run-time details of the jobs. The
// availability of the values depends on the underlying DRMAA2 implementation
// of the execution Context.
func (j *Job) JobInfos() []drmaa2interface.JobInfo {
	jis := make([]drmaa2interface.JobInfo, 0, len(j.tasklist))
	for _, task := range j.tasklist {
		if task.job != nil {
			ji, err := task.job.GetJobInfo()
			if err != nil {
				continue
			}
			jis = append(jis, ji)
		}
	}
	return jis
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

// Run submits a task which executes the given command and args. The command
// needs to be available on the execution backend.
func (j *Job) Run(cmd string, args ...string) *Job {
	jt := drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args}
	return j.RunT(jt)
}

// RunT submits a task given specified with the JobTemplate.
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
	jobTemplate, _ := copystructure.Copy(jt)
	j.tasklist = append(j.tasklist, &task{job: job, submitError: err,
		template: jobTemplate.(drmaa2interface.JobTemplate)})
	return j
}

// Do executes a function which gets the DRMAA2 job object as parameter.
// This allows working with the low-level DRMAA2 job object.
func (j *Job) Do(f func(job drmaa2interface.Job)) *Job {
	if job, err := j.jobCheck(); err != nil {
		// do not store error as it overrides job action errors
		return j
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

// LastError returns the error if occurred during last job operation.
func (j *Job) LastError() error {
	return j.lastError
}

func rerunTask(j *Job, e *task) {
	job, err := j.wfl.js.RunJob(e.template)
	j.lastError = err
	if err == nil {
		jobTemplate, _ := copystructure.Copy(e.template)
		j.tasklist = append(j.tasklist, &task{job: job, submitError: err,
			template: jobTemplate.(drmaa2interface.JobTemplate)})
	}
}

func replaceTask(j *Job, e *task) {
	e.job, e.submitError = j.wfl.js.RunJob(e.template)
}

// Resubmit starts the previously submitted job n-times. The jobs are
// executed in parallel.
func (j *Job) Resubmit(r int) *Job {
	for i := 0; i < r || r == -1; i++ {
		if e := j.lastJob(); e != nil {
			rerunTask(j, e)
		} else {
			j.lastError = errors.New("job not available")
			break
		}
	}
	return j
}

// AnyFailed returns true when at least job in the whole chain failed.
func (j *Job) AnyFailed() bool {
	for _, task := range j.tasklist {
		if task.job.GetState() == drmaa2interface.Failed {
			return true
		}
	}
	return false
}

// Blocking

// RunEvery provides the same functionally like RunEveryT but the job is created
// based on the given command with the arguments.
func (j *Job) RunEvery(d time.Duration, end time.Time, cmd string, args ...string) error {
	return j.RunEveryT(d, end, drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// RunEveryT submits a job every d time.Duration regardless if the previously
// job is still running or finished or failed. The method only aborts and returns
// an error if an error during job submission happened and the job could not
// be submitted.
func (j *Job) RunEveryT(d time.Duration, end time.Time, jt drmaa2interface.JobTemplate) error {
	for range time.NewTicker(d).C {
		if time.Now().After(end) {
			return nil
		}
		j.RunT(jt)
		if j.lastError != nil {
			return j.lastError
		}
	}
	return nil
}

// After blocks the given duration and continues by returning the same job.
func (j *Job) After(d time.Duration) *Job {
	<-time.After(d)
	return j
}

func wait(task *task) {
	if task.job == nil {
		return
	}
	task.terminationError = task.job.WaitTerminated(drmaa2interface.InfiniteTime)
	task.terminated = true
	task.jobinfo, task.jobinfoError = task.job.GetJobInfo()
}

// Wait until the most recently job was finished.
func (j *Job) Wait() *Job {
	j.lastError = nil
	if task := j.lastJob(); task != nil {
		wait(task)
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

// Synchronize with all jobs in the chain. All jobs are terminated when
// the call returns.
func (j *Job) Synchronize() *Job {
	for _, task := range j.tasklist {
		wait(task)
	}
	return j
}

// ListAllFailed returns all jobs which failed. Note that it implicitly
// waits until all tasks finished.
func (j *Job) ListAllFailed() []drmaa2interface.Job {
	failed := make([]drmaa2interface.Job, 0, len(j.tasklist))
	for _, task := range j.tasklist {
		wait(task)
		if task.job.GetState() == drmaa2interface.Failed {
			failed = append(failed, task.job)
		}
	}
	return failed
}

// HasAnyFailed returns true if there is any failed task in the chain.
// Note that the functions implicitly waits until all tasks finsihed.
func (j *Job) HasAnyFailed() bool {
	failed := j.ListAllFailed()
	return len(failed) == 0
}

// RetryAnyFailed reruns any failed tasks in the job and replaces them
// with the new incarnation.
func (j *Job) RetryAnyFailed(amount int) *Job {
	for i := 0; i < amount || amount == -1; i++ {
		for _, task := range j.tasklist {
			wait(task)
			if task.job.GetState() == drmaa2interface.Failed {
				replaceTask(j, task)
			}
		}
		if !j.HasAnyFailed() {
			break
		}
	}
	return j
}

// Failed returns true in case the current job stated equals drmaa2interface.Failed
func (j *Job) Failed() bool {
	return j.State() == drmaa2interface.Failed
}

// Success returns true in case the current job stated equals drmaa2interface.Done
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
	if task := j.lastJob(); task != nil {
		return task.jobinfo.ExitStatus
	}
	return -1
}

// Then waits until the previous job is terminated and executes the
// given function by providing the DRMAA2 job interface which gives
// access to the low-level DRMAA2 job.
func (j *Job) Then(f func(job drmaa2interface.Job)) *Job {
	j.lastError = nil
	if task := j.lastJob(); task != nil && task.job != nil {
		task.terminationError = task.job.WaitTerminated(drmaa2interface.InfiniteTime)
		task.terminated = true
		task.jobinfo, task.jobinfoError = task.job.GetJobInfo()
		f(task.job)
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

// ThenRunT waits until the previous job is terminated and executes then
// a new job based on the given JobTemplate.
func (j *Job) ThenRunT(jt drmaa2interface.JobTemplate) *Job {
	return j.Wait().RunT(jt)
}

func waitForJobEndAndState(j *Job) drmaa2interface.JobState {
	job, err := j.jobCheck()
	if err != nil {
		return drmaa2interface.Undetermined
	}
	lastError := job.WaitTerminated(drmaa2interface.InfiniteTime)
	if lastError != nil {
		return drmaa2interface.Undetermined
	}
	return job.GetState()
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
	return j.OnSuccessRunT(drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// OnSuccessRunT submits a job when the previous job ended in the
// job state drmaa2interface.Done.
func (j *Job) OnSuccessRunT(jt drmaa2interface.JobTemplate) *Job {
	if waitForJobEndAndState(j) == drmaa2interface.Done {
		j.RunT(jt)
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
	return j.OnFailureRunT(drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// OnFailureRunT submits a job when the previous job ended in a state
// different than drmaa2interface.Done.
func (j *Job) OnFailureRunT(jt drmaa2interface.JobTemplate) *Job {
	if waitForJobEndAndState(j) != drmaa2interface.Done {
		j.RunT(jt)
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
