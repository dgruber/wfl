package wfl

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"os"
)

// Observer is a collection of functions which implements
// behavior which should be executed when a task submission
// failed, when the task failed, or when then the job was
// running successfully.
type Observer struct {
	ErrorHandler   func(error)
	FailedHandler  func(drmaa2interface.Job)
	SuccessHandler func(drmaa2interface.Job)
}

// NewDefaultObserver returns an Observer which panics when
// a task submission error occurred, prints a message and exits
// the application when the task exits with error code != 0,
// and prints a message and continues when a task was running
// successfully.
func NewDefaultObserver() Observer {
	return Observer{
		ErrorHandler: func(e error) { panic(e) },
		FailedHandler: func(j drmaa2interface.Job) {
			fmt.Printf("job %s failed\n", j.GetID())
			os.Exit(1)
		},
		SuccessHandler: func(j drmaa2interface.Job) {
			fmt.Printf("job %s finished successfully\n", j.GetID())
		},
	}
}

// Observe executes the functions defined in the Observer
// when task submission errors, the task failed, and
// when the job finished successfully. Note that this is
// a blocking call.
func (j *Job) Observe(o Observer) *Job {
	return j.OnError(o.ErrorHandler).
		OnFailure(o.FailedHandler).
		OnSuccess(o.SuccessHandler)
}
