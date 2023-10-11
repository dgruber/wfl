package wfl

import (
	"context"
	"errors"
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl/pkg/log"
)

// Workflow contains the backend context and a job session. The DRMAA2 job
// session provides typically logical isolation between jobs.
type Workflow struct {
	ctx                   *Context
	js                    drmaa2interface.JobSession
	workflowCreationError error
	log                   log.Logger
	llmConfig             *llmConfig
}

// NewWorkflow creates a new Workflow based on the given execution context.
// Internally it creates a DRMAA2 JobSession which is used for separating jobs.
func NewWorkflow(context *Context) *Workflow {
	var err error
	logger, _ := log.NewZerologger()
	if context == nil {
		err = errors.New("no context given")
	} else if context.SM == nil {
		err = errors.New("no Session Manager available in context")
	} else {
		if context.JobSessionName == "" {
			context.JobSessionName = "wfl"
		}
		js, errJS := context.SM.CreateJobSession(context.JobSessionName, "")
		if errJS != nil {
			var errOpenJS error
			if js, errOpenJS = context.SM.OpenJobSession(context.JobSessionName); errOpenJS != nil {
				err = fmt.Errorf("error creating (%v) or opening (%v) Job Session \""+
					context.JobSessionName+"\"", errJS, errOpenJS)
			}
		}
		return &Workflow{ctx: context,
			js:                    js,
			workflowCreationError: err,
			log:                   logger,
		}
	}
	return &Workflow{ctx: nil,
		workflowCreationError: err,
		log:                   logger,
	}
}

// Logger return the current logger of the workflow.
func (w *Workflow) Logger() log.Logger {
	return w.log
}

// SetLogger sets a new logger for the workflow which writes
// processes internal log messages. Note that nil loggers are
// not accepted.
//
// Example: w.SetLogger(log.NewKlogLogger("INFO"))
func (w *Workflow) SetLogger(log log.Logger) *Workflow {
	if log != nil {
		w.log = log
	}
	return w
}

// SetLogLevel changes the log level filter for the workflow.
// The default log level is log.Warning.
func (w *Workflow) SetLogLevel(logLevel log.LogLevel) *Workflow {
	w.log.SetLogLevel(logLevel)
	return w
}

// OnError executes a function if happened during creating a job session
// or opening a job session.
func (w *Workflow) OnError(f func(e error)) *Workflow {
	if w.workflowCreationError != nil {
		f(w.workflowCreationError)
	}
	return w
}

// OnErrorPanic panics if happened during creating a job session
// or opening a job session.
func (w *Workflow) OnErrorPanic() *Workflow {
	if w.workflowCreationError != nil {
		panic(w.workflowCreationError)
	}
	return w
}

// Error returns the error if happened during creating a job session
// or opening a job session.
func (w *Workflow) Error() error {
	return w.workflowCreationError
}

// HasError returns true if there was an error during creating a job session
// or opening a job session.
func (w *Workflow) HasError() bool {
	return w.workflowCreationError != nil
}

// Run submits the first task in the workflow and returns the Job object.
// Same as NewJob(w).Run().
func (w *Workflow) Run(cmd string, args ...string) *Job {
	job := NewJob(w)
	// as we call Run() from workflow wrapper we need to set
	// logging depth for indicating calling function name
	// one level deeper
	job.ctx = context.WithValue(job.ctx, "log-depth", 4)
	job.Run(cmd, args...)
	job.ctx = context.WithValue(job.ctx, "log-depth", 3)
	return job
}

// RunT submits the first task in the workflow and returns the Job object.
// Same as NewJob(w).RunT().
func (w *Workflow) RunT(jt drmaa2interface.JobTemplate) *Job {
	job := NewJob(w)
	job.ctx = context.WithValue(job.ctx, "log-depth", 4)
	job.RunT(jt)
	job.ctx = context.WithValue(job.ctx, "log-depth", 3)
	return job
}

// RunArrayJob executes the given command multiple times as specified with begin,
// end, and step. To run a command 10 times, begin can be set to 1, end to 10 and
// step to 1. maxParallel can limit the amount of executions which are running in
// parallel if supported by the context. The process context sets the TASK_ID env
// variable to the task ID.
func (w *Workflow) RunArrayJob(begin, end, step, maxParallel int, cmd string, args ...string) *Job {
	job := NewJob(w)
	// klog prints the line of code which called RunArrayJob()
	job.ctx = context.WithValue(job.ctx, "log-depth", 4)
	job.RunArray(begin, end, step, maxParallel, cmd, args...)
	job.ctx = context.WithValue(job.ctx, "log-depth", 3)
	return job
}

// RunArrayJob executes the given job defined in the JobTemplate multiple times.
// See RunArrayJob().
func (w *Workflow) RunArrayJobT(begin, end, step, maxParallel int, jt drmaa2interface.JobTemplate) *Job {
	job := NewJob(w)
	// klog prints the line of code which called RunArrayJobT()
	job.ctx = context.WithValue(job.ctx, "log-depth", 4)
	job.RunArrayT(begin, end, step, maxParallel, jt)
	job.ctx = context.WithValue(job.ctx, "log-depth", 3)
	return job
}

func (w *Workflow) RunMatrixT(jt drmaa2interface.JobTemplate, x, y Replacement) *Job {
	job := NewJob(w)
	// klog prints the line of code which called RunArrayJobT()
	job.ctx = context.WithValue(job.ctx, "log-depth", 4)
	job.RunMatrixT(jt, x, y)
	job.ctx = context.WithValue(job.ctx, "log-depth", 3)
	return job
}

// ListJobs returns all jobs visible in the workflow (i.e. available
// in the underlying drmaa2session). It may wrap one task in one Job
// object and return multiple Job objects even when only one Job with
// many tasks was submitted.
func (w *Workflow) ListJobs() []*Job {
	d2jobs, err := w.js.GetJobs(drmaa2interface.CreateJobInfo())
	if err != nil {
		w.log.Errorf(context.Background(), "workflow.ListJobs() failed with: %v", err)
		return nil
	}
	jobs := make([]*Job, 0, len(d2jobs))

	for _, d2job := range d2jobs {
		template, _ := d2job.GetJobTemplate()
		jobInfo, _ := d2job.GetJobInfo()
		state := d2job.GetState()
		terminated := false
		if state == drmaa2interface.Failed || state == drmaa2interface.Done {
			terminated = true
		}
		jobs = append(jobs,
			&Job{
				ctx: context.Background(),
				wfl: w,
				tasklist: []*task{{
					job:                             d2job,
					template:                        template,
					jobinfo:                         jobInfo,
					terminated:                      terminated,
					submitError:                     nil,
					terminationError:                nil,
					jobinfoError:                    nil,
					retry:                           0,
					waitForEndStateCollectedJobInfo: false,
					isJobArray:                      false,
					jobArray:                        nil,
				}},
			})
	}

	return jobs
}

// NewJob creates a new empty Job object for the given workflow.
// Equivalent to NewJob(*Workflow).
func (w *Workflow) NewJob() *Job {
	return NewJob(w)
}
