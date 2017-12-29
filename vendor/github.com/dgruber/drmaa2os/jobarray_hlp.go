package drmaa2os

import (
	"errors"
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
	var globalError string
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
			globalError = fmt.Sprintf("Job %s error: %s %s", jobs[i].GetID(), err, globalError)
		}
	}
	if globalError == "" {
		return nil
	}
	return errors.New(globalError)
}
