package wfl

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
)

// Workflow contains the backend context and a job session. The DRMAA2 job session
// provides typically logical isolation between jobs.
type Workflow struct {
	ctx                   *Context
	js                    drmaa2interface.JobSession
	workflowCreationError error
}

// NewWorkflow creates a new Workflow based on the given execution context.
// Internally it creates a DRMAA2 JobSession which is used for separating jobs.
func NewWorkflow(context *Context) *Workflow {
	var err error
	if context == nil {
		err = errors.New("No context given")
	} else if context.sm == nil {
		err = errors.New("No Session Manager available in context")
	} else {
		js, errJS := context.sm.CreateJobSession("wfl", "")
		if errJS != nil {
			var errOpenJS error
			if js, errOpenJS = context.sm.OpenJobSession("wfl"); errOpenJS != nil {
				err = fmt.Errorf("Error creating (%s) or opening (%s) Job Session \"wfl\"\n", errJS.Error(), errOpenJS.Error())
			}
		}
		return &Workflow{ctx: context, js: js, workflowCreationError: err}
	}
	return &Workflow{workflowCreationError: err, ctx: nil}
}

// OnError executes a function if happened during creating a job session
// or opening a job session.
func (w *Workflow) OnError(f func(e error)) *Workflow {
	if w.workflowCreationError != nil {
		f(w.workflowCreationError)
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
	return NewJob(w).Run(cmd, args...)
}

// RunT submits the first task in the workflow and returns the Job object.
// Same as NewJob(w).RunT().
func (w *Workflow) RunT(jt drmaa2interface.JobTemplate) *Job {
	return NewJob(w).RunT(jt)
}
