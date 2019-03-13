package drmaa2os

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
)

type action int

const (
	suspend action = iota
	resume
	hold
	release
	terminate
)

func jobAction(a action, jobs []drmaa2interface.Job) error {
	var globalError error
	for i := range jobs {
		var err error
		switch a {
		case suspend:
			err = jobs[i].Suspend()
		case resume:
			err = jobs[i].Resume()
		case hold:
			err = jobs[i].Hold()
		case release:
			err = jobs[i].Release()
		case terminate:
			err = jobs[i].Terminate()
		}
		if err != nil {
			if globalError != nil {
				globalError = fmt.Errorf("Job %s error: %s | %s",
					jobs[i].GetID(), err, globalError.Error())
			} else {
				globalError = fmt.Errorf("Job %s error: %s",
					jobs[i].GetID(), err)
			}
		}
	}
	return globalError
}

func newArrayJob(id, jsessionname string, tmpl drmaa2interface.JobTemplate, jobs []drmaa2interface.Job) *ArrayJob {
	return &ArrayJob{
		id:          id,
		sessionname: jsessionname,
		template:    tmpl,
		jobs:        jobs,
	}
}
