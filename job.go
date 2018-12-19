package wfl

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
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
	ctx       context.Context // logging
}

// NewJob creates the initial empty job with the given workflow.
func NewJob(wfl *Workflow) *Job {
	return &Job{
		wfl:      wfl,
		tasklist: make([]*task, 0, 32),
		ctx:      context.Background(),
	}
}

// EmptyJob creates an empty job.
func EmptyJob() *Job {
	return &Job{}
}

// Job Sequence Properties

// TagWith tags a job with a string for identification. Global for all tasks of the job.
func (j *Job) TagWith(tag string) *Job {
	j.begin(j.ctx, fmt.Sprintf("TagWith(%s)", tag))
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
	j.begin(j.ctx, "Template()")
	j.lastError = nil
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		if template, errTmp := job.GetJobTemplate(); errTmp != nil {
			j.errorf(j.ctx, "Template() [JobID: %s]: GetJobTemplate() failed with %s",
				j.JobID(), errTmp.Error())
			j.lastError = errTmp
		} else {
			return &template
		}
	}
	return nil
}

// State returns the current state of the job previously submitted.
func (j *Job) State() drmaa2interface.JobState {
	j.begin(j.ctx, "State()")
	job, err := j.jobCheck()
	if err != nil {
		j.lastError = err
		return drmaa2interface.Undetermined
	}
	return job.GetState()
}

// JobID returns the job ID of the previously submitted job.
func (j *Job) JobID() string {
	j.begin(j.ctx, "JobID()")
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
	j.begin(j.ctx, "JobInfo()")
	job, err := j.jobCheck()
	if err != nil {
		j.lastError = err
		return drmaa2interface.JobInfo{}
	}
	ji, errJI := job.GetJobInfo()
	if errJI != nil {
		j.errorf(j.ctx, "JobInfo() [JobID: %s]: GetJobInfo() failed with: %s",
			j.JobID(), errJI.Error())
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
	j.begin(j.ctx, "JobInfos()")
	jis := make([]drmaa2interface.JobInfo, 0, len(j.tasklist))
	for _, task := range j.tasklist {
		if task.job != nil {
			ji, err := task.job.GetJobInfo()
			if err != nil {
				j.warningf(j.ctx,
					"task returned error when calling GetJobInfo(): %s",
					err.Error())
				continue
			}
			jis = append(jis, ji)
		}
	}
	return jis
}

// Run submits a task which executes the given command and args. The command
// needs to be available on the execution backend.
func (j *Job) Run(cmd string, args ...string) *Job {
	j.begin(j.ctx, fmt.Sprintf("Run(%s, %v)", cmd, args))
	jt := drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args}
	return j.RunT(jt)
}

// RunT submits a task given specified with the JobTemplate.
func (j *Job) RunT(jt drmaa2interface.JobTemplate) *Job {
	j.begin(j.ctx, fmt.Sprintf("RunT(%s, %v)", jt.RemoteCommand, jt.Args))
	if err := j.checkCtx(); err != nil {
		j.lastError = err
		return j
	}
	// merging only specific job template parameters
	jt = mergeJobTemplateWithDefaultTemplate(jt, j.wfl.ctx.defaultTemplate)

	// JobCategory overrides all at the moment...
	if jt.JobCategory == "" {
		jt.JobCategory = j.wfl.ctx.defaultDockerImage
	}
	if j.wfl.js == nil {
		j.lastError = errors.New("JobSession is nil")
		return j
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
	j.begin(j.ctx, fmt.Sprintf("Do(%s)",
		runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()),
	)
	job, err := j.jobCheck()
	// do not store error as it overrides job action errors
	if err == nil {
		f(job)
	} else {
		j.errorf(j.ctx,
			"Do(): Function (%s) is not executed as task is nil",
			runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(),
		)
	}
	return j
}

// Suspend stops the last task of the job from execution. How this is
// done depends on the Context. Typically a signal (like SIGTSTP) is
// sent to the tasks of the job.
func (j *Job) Suspend() *Job {
	j.begin(j.ctx, "Suspend()")
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		j.lastError = job.Suspend()
	}
	return j
}

// Resume continues a suspended job to continue execution.
func (j *Job) Resume() *Job {
	j.begin(j.ctx, "Resume()")
	if job, err := j.jobCheck(); err != nil {
		j.lastError = err
	} else {
		j.lastError = job.Resume()
	}
	return j
}

// Kill stops the job from execution.
func (j *Job) Kill() *Job {
	j.begin(j.ctx, "Kill()")
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

// Resubmit starts the previously submitted task n-times. All tasks are
// executed in parallel.
func (j *Job) Resubmit(r int) *Job {
	j.begin(j.ctx, fmt.Sprintf("Resubmit(%d)", r))
	for i := 0; i < r || r == -1; i++ {
		if t := j.lastJob(); t != nil {
			rerunTask(j, t)
		} else {
			j.errorf(
				j.ctx,
				"Resubmit(): Could not find any job in order to re-run it.",
			)
			j.lastError = errors.New("job not available")
			break
		}
	}
	return j
}

// AnyFailed returns true when at least job in the whole chain failed.
func (j *Job) AnyFailed() bool {
	j.begin(j.ctx, "AnyFailed()")
	for _, task := range j.tasklist {
		if task.job.GetState() == drmaa2interface.Failed {
			return true
		}
	}
	return false
}

// RunEvery provides the same functionally like RunEveryT but the job is created
// based on the given command with the arguments.
func (j *Job) RunEvery(d time.Duration, end time.Time, cmd string, args ...string) error {
	j.begin(j.ctx, fmt.Sprintf("RunEvery(%s %s %s %s)",
		d.String(),
		end.Format("15:04:05"),
		cmd,
		args),
	)
	return j.RunEveryT(d, end, drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// RunEveryT submits a job every d time.Duration regardless if the previously
// job is still running or finished or failed. The method only aborts and returns
// an error if an error during job submission happened and the job could not
// be submitted.
func (j *Job) RunEveryT(d time.Duration, end time.Time, jt drmaa2interface.JobTemplate) error {
	j.begin(j.ctx, fmt.Sprintf("RunEvery(%s %s %s %s)",
		d.String(),
		end.Format("15:04:05"),
		jt.RemoteCommand,
		jt.Args),
	)
	for range time.NewTicker(d).C {
		if time.Now().After(end) {
			j.infof(j.ctx, "RunEveryT() end time reached")
			break
		}
		j.infof(j.ctx, "RunEveryT() submit job")
		j.RunT(jt)
		if j.lastError != nil {
			j.errorf(
				j.ctx,
				"RunEveryT: Aborting: Job submission failed for job %s with %s",
				j.JobID(),
				j.lastError.Error(),
			)
			return j.lastError
		}
	}
	return nil
}

// After blocks the given duration and continues by returning the same job.
func (j *Job) After(d time.Duration) *Job {
	j.infof(j.ctx, "After()")
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

// Wait until the most recently task was finished.
func (j *Job) Wait() *Job {
	j.infof(j.ctx, "Wait()")
	j.lastError = nil
	if task := j.lastJob(); task != nil {
		wait(task)
	} else {
		j.errorf(
			j.ctx,
			"Wait() has no task to wait for",
		)
		j.lastError = errors.New("task not available")
	}
	return j
}

// Retry waits until the last task in chain (not for the previous ones) is finished.
// When it failed it resubmits it and waits again for a successful end.
func (j *Job) Retry(r int) *Job {
	j.infof(j.ctx, "Retry()")
	for ; r > 0; r-- {
		if j.Wait().Success() {
			j.infof(j.ctx, "Retry(): Last task run successfully. No restart required.")
			return j
		}
		j.warningf(j.ctx, "Retry(): Last task failed. Resubmitting task %s.", j.JobID())
		j.Resubmit(1)
	}
	return j
}

// Synchronize waits until the tasks of the job are finished. All jobs are terminated when
// the call returns.
func (j *Job) Synchronize() *Job {
	j.begin(j.ctx, "Synchronize()")
	for _, task := range j.tasklist {
		wait(task)
	}
	return j
}

// ListAllFailed returns all tasks which failed as array of DRMAA2 jobs. Note that
// it implicitly waits until all tasks are finished.
func (j *Job) ListAllFailed() []drmaa2interface.Job {
	j.begin(j.ctx, "ListAllFailed()")
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
	j.begin(j.ctx, "HasAnyFailed()")
	failed := j.ListAllFailed()
	return len(failed) > 0
}

// RetryAnyFailed reruns any failed tasks and replaces them
// with a new task incarnation.
func (j *Job) RetryAnyFailed(amount int) *Job {
	j.begin(j.ctx, fmt.Sprintf("RetryAnyFailed(%d)", amount))
	for i := 0; i < amount || amount == -1; i++ {
		for _, task := range j.tasklist {
			wait(task)
			if task.job.GetState() == drmaa2interface.Failed {
				failedJobID := task.job.GetID()
				replaceTask(j, task)
				j.warningf(j.ctx, "RetryAnyFailed(%d)): Task %s failed. Retry task (%s).",
					amount, failedJobID, task.job.GetID())
			}
		}
		if !j.HasAnyFailed() {
			break
		}
	}
	return j
}

// Failed returns true in case the current task stated equals drmaa2interface.Failed
func (j *Job) Failed() bool {
	return j.State() == drmaa2interface.Failed
}

// Success returns true in case the current task stated equals drmaa2interface.Done
// and the job exit status is 0.
func (j *Job) Success() bool {
	if j.State() == drmaa2interface.Done {
		if j.ExitStatus() == 0 {
			return true
		}
	}
	return false
}

// Errored returns if an error occurred at the last operation.
func (j *Job) Errored() bool {
	if j.lastError != nil {
		return true
	}
	return false
}

// ExitStatus waits until the previously submitted task is finished and
// returns the exit status of the task. In case of an internal error it
// returns -1.
func (j *Job) ExitStatus() int {
	j.infof(j.ctx, "ExitStatus()")
	j.Wait()
	if task := j.lastJob(); task != nil {
		return task.jobinfo.ExitStatus
	}
	j.errorf(j.ctx, "ExitStatus(): task not found")
	return -1
}

// Then waits until the previous task is terminated and executes the
// given function by providing the DRMAA2 job interface which gives
// access to the low-level DRMAA2 job methods.
func (j *Job) Then(f func(job drmaa2interface.Job)) *Job {
	j.begin(j.ctx, fmt.Sprintf("Then(%s)",
		runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()))
	j.lastError = nil
	if task := j.lastJob(); task != nil && task.job != nil {
		task.terminationError = task.job.WaitTerminated(drmaa2interface.InfiniteTime)
		task.terminated = true
		task.jobinfo, task.jobinfoError = task.job.GetJobInfo()
		f(task.job)
	} else {
		j.errorf(j.ctx, "Then(%s): task not found",
			runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		j.lastError = errors.New("task not available")
	}
	return j
}

// ThenRun waits until the previous task is terminated and executes then
// the given command as new task.
func (j *Job) ThenRun(cmd string, args ...string) *Job {
	j.begin(j.ctx, "ThenRun()")
	return j.Wait().Run(cmd, args...)
}

// ThenRunT waits until the previous task is terminated and executes then
// a new task based on the given JobTemplate.
func (j *Job) ThenRunT(jt drmaa2interface.JobTemplate) *Job {
	j.begin(j.ctx, "ThenRunT()")
	return j.Wait().RunT(jt)
}

// OnSuccess executes the given function after the previously submitted
// task finished in the drmaa2interface.Done state.
func (j *Job) OnSuccess(f func(job drmaa2interface.Job)) *Job {
	if waitForJobEndAndState(j) == drmaa2interface.Done {
		j.Then(f)
	}
	return j
}

// OnSuccessRun submits a task when the previous task ended in the
// state drmaa2interface.Done.
func (j *Job) OnSuccessRun(cmd string, args ...string) *Job {
	j.begin(j.ctx, fmt.Sprintf("OnSuccessRun(%s %v)", cmd, args))
	return j.OnSuccessRunT(drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// OnSuccessRunT submits a task when the previous task ended in the
// state drmaa2interface.Done.
func (j *Job) OnSuccessRunT(jt drmaa2interface.JobTemplate) *Job {
	j.begin(j.ctx, "OnSuccessRunT()")
	if waitForJobEndAndState(j) == drmaa2interface.Done {
		j.infof(j.ctx, "OnSuccessRunT(): Previous task run successfully. Running new task.")
		j.RunT(jt)
	}
	return j
}

// OnFailure executes the given function when the previous task in the list failed.
// Fails mean the job was started successfully by the system but then existed with
// an exit code != 0.
//
// When running the task resulted in an error (i.e. the job run function errored),
// then the function is not executed.
func (j *Job) OnFailure(f func(job drmaa2interface.Job)) *Job {
	j.begin(j.ctx, fmt.Sprintf("OnFailure(%s)",
		runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()),
	)
	if waitForJobEndAndState(j) != drmaa2interface.Done {
		j.infof(j.ctx, "OnFailure(%s): Previous task failed. Executing function.",
			runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
		j.Then(f)
	}
	return j
}

// OnFailureRun submits a task when the previous task ended in a state
// different than drmaa2interface.Done.
func (j *Job) OnFailureRun(cmd string, args ...string) *Job {
	j.begin(j.ctx, fmt.Sprintf("OnFailureRun(%s %v)",
		cmd, args))
	return j.OnFailureRunT(drmaa2interface.JobTemplate{RemoteCommand: cmd, Args: args})
}

// OnFailureRunT submits a task when the previous job ended in a state
// different than drmaa2interface.Done.
func (j *Job) OnFailureRunT(jt drmaa2interface.JobTemplate) *Job {
	j.begin(j.ctx, "OnFailureRunT()")
	if waitForJobEndAndState(j) != drmaa2interface.Done {
		j.RunT(jt)
	}
	return j
}

// OnError executes the given function if the last Job operation resulted
// in an error (like a job submission failure).
func (j *Job) OnError(f func(err error)) *Job {
	j.begin(j.ctx, fmt.Sprintf("OnError(%s)",
		runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()),
	)
	if j.lastError != nil {
		f(j.lastError)
	}
	return j
}
