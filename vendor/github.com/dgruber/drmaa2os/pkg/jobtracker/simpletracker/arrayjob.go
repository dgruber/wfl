package simpletracker

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
)

// arrayJobSubmissionController starts and supervises all jobs of a job array.
// It takes care that not more jobs than _maxParallel_ jobs are running
// at the same time. When jobs are finished it starts more jobs and
// put their state from _queued_ into _running_ state.
func arrayJobSubmissionController(jt *JobTracker, arrayjobid string, t drmaa2interface.JobTemplate,
	begin, end, step, maxParallel int) chan error {
	firstJobErrorCh := make(chan error, 1)
	go func() {
		waitCh := make(chan int, maxParallel)
		for i := begin; i <= end; i += step {
			if maxParallel > 0 {
				waitCh <- i // block when buffer is full - wait until jobs are finished
			}
			jobid := fmt.Sprintf("%s.%d", arrayjobid, i)
			jt.ps.Lock()
			// check if job was cancelled while waiting
			if jt.ps.jobState[jobid] == drmaa2interface.Failed {
				jt.ps.Unlock()
				<-waitCh
				continue
			}
			jt.ps.Unlock()
			pid, err := StartProcess(jobid, i, t, jt.ps.jobch)
			if err != nil {
				// job failed
				jt.ps.Lock()
				jt.ps.jobState[jobid] = drmaa2interface.Failed
				jt.ps.Unlock()
				if i == begin {
					firstJobErrorCh <- err
				}
				<-waitCh
				continue
			}

			if i == begin {
				firstJobErrorCh <- nil
			}
			jt.Lock()
			jt.js.SaveArrayJobPID(arrayjobid, i, pid)
			jt.Unlock()
			if maxParallel > 0 {
				go func() {
					jt.Wait(jobid, 0.0, drmaa2interface.Done, drmaa2interface.Failed)
					<-waitCh
				}()
			}
		}
	}()
	return firstJobErrorCh
}
