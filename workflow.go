package wfl

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
)

type Workflow struct {
	ctx                   *Context
	js                    drmaa2interface.JobSession
	workflowCreationError error
	jobs                  []*Job
}

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
		return &Workflow{ctx: context, js: js, workflowCreationError: err, jobs: make([]*Job, 0, 1)}
	}
	return &Workflow{workflowCreationError: err, ctx: nil}
}

func (w *Workflow) OnError(f func(e error)) *Workflow {
	if w.workflowCreationError != nil {
		f(w.workflowCreationError)
	}
	return w
}

func (w *Workflow) Error() error {
	return w.workflowCreationError
}

func (w *Workflow) HasError() bool {
	if w.workflowCreationError != nil {
		return true
	}
	return false
}

func (w *Workflow) Run(cmd string, args ...string) *Job {
	return NewJob(w).Run(cmd, args...)
}

func (w *Workflow) RunT(jt drmaa2interface.JobTemplate) *Job {
	return NewJob(w).RunT(jt)
}
