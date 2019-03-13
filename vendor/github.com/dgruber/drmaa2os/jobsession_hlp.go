package drmaa2os

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

func newJobSession(name string, tracker []jobtracker.JobTracker) *JobSession {
	return &JobSession{
		name:    name,
		tracker: tracker,
	}
}

func waitAny(waitForStartedState bool, jobs []drmaa2interface.Job, timeout time.Duration) (drmaa2interface.Job, error) {
	started := make(chan int, len(jobs))
	errored := make(chan int, len(jobs))
	abort := make(chan bool, len(jobs))

	for i := 0; i < len(jobs); i++ {
		index := i // closure fun
		job := jobs[i]
		waitForStarted := waitForStartedState
		go func() {
			finished := make(chan bool, 1)
			go func() {
				var errWait error
				if waitForStarted {
					errWait = job.WaitStarted(timeout)
				} else {
					errWait = job.WaitTerminated(timeout)
				}
				if errWait == nil {
					started <- index
				} else {
					errored <- index
				}
				finished <- true
				return
			}()
			select {
			case <-abort:
				return
			case <-finished:
				return
			}
		}()
	}

	timeoutCh := time.Tick(timeout)
	errorCnt := 0

	for {
		select {
		case <-errored:
			errorCnt++
			if errorCnt >= len(jobs) {
				return nil, errors.New("Error waiting for jobs")
			}
			continue
		case jobindex := <-started:
			// abort all waiting go routines
			for i := 1; i <= len(jobs)-errorCnt; i++ {
				abort <- true
			}
			return jobs[jobindex], nil
		case <-timeoutCh:
			return nil, ErrorInvalidState
		}
	}
}
