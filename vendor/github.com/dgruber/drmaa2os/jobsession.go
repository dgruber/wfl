package drmaa2os

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/d2hlp"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

type JobSession struct {
	name    string
	tracker []jobtracker.JobTracker
}

func NewJobSession(name string, tracker []jobtracker.JobTracker) *JobSession {
	return &JobSession{
		name:    name,
		tracker: tracker,
	}
}

func (js *JobSession) Close() error {
	js.name = ""
	js.tracker = nil
	return nil
}

func (js *JobSession) GetContact() (string, error) {
	return "", nil
}

func (js *JobSession) GetSessionName() (string, error) {
	return js.name, nil
}

func (js *JobSession) GetJobCategories() ([]string, error) {
	return []string{}, nil
}

func createJobFromInfo(ji drmaa2interface.JobInfo) Job {
	return Job{}
}

func (js *JobSession) GetJobs(filter drmaa2interface.JobInfo) ([]drmaa2interface.Job, error) {
	var joblist []drmaa2interface.Job

	for _, tracker := range js.tracker {
		jobs, err := tracker.ListJobs()
		if err != nil {
			return nil, err
		}
		for _, jobid := range jobs {
			if jinfo, err := tracker.JobInfo(jobid); err != nil {
				continue
			} else {
				if d2hlp.JobInfoMatches(jinfo, filter) {
					// TODO get template from DB
					jobtemplate := drmaa2interface.JobTemplate{}

					job := NewJob(jobid, js.name, jobtemplate, tracker)
					joblist = append(joblist, drmaa2interface.Job(job))
				}
			}
		}
	}

	return joblist, nil
}

func (js *JobSession) GetJobArray(id string) (drmaa2interface.ArrayJob, error) {
	jobids, err := js.tracker[0].ListArrayJobs(id)
	if err != nil {
		return nil, err
	}
	joblist := make([]drmaa2interface.Job, 0, len(jobids))
	for _, id := range jobids {
		// TODO get template from DB
		jobtemplate := drmaa2interface.JobTemplate{}

		job := NewJob(id, js.name, jobtemplate, js.tracker[0])
		joblist = append(joblist, drmaa2interface.Job(job))
	}
	return NewArrayJob(id, js.name, drmaa2interface.JobTemplate{}, joblist), nil
}

func (js *JobSession) RunJob(jt drmaa2interface.JobTemplate) (drmaa2interface.Job, error) {
	id, err := js.tracker[0].AddJob(jt)
	if err != nil {
		return nil, err
	}
	return NewJob(id, js.name, jt, js.tracker[0]), nil
}

func (js *JobSession) RunBulkJobs(jt drmaa2interface.JobTemplate, begin int, end int, step int, maxParallel int) (drmaa2interface.ArrayJob, error) {
	id, err := js.tracker[0].AddArrayJob(jt, begin, end, step, maxParallel)
	if err != nil {
		return nil, err
	}
	return js.GetJobArray(id)
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
			select {
			case <-abort:
				fmt.Println("abort")
				return
			default:
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
				abort <- true // TODO multiple abort (len - errCnt)
			}
			return jobs[jobindex], nil
		case <-timeoutCh:
			return nil, ErrorInvalidState
		}
	}
}

func (js *JobSession) WaitAnyStarted(jobs []drmaa2interface.Job, timeout time.Duration) (drmaa2interface.Job, error) {
	return waitAny(true, jobs, timeout)
}

func (js *JobSession) WaitAnyTerminated(jobs []drmaa2interface.Job, timeout time.Duration) (drmaa2interface.Job, error) {
	return waitAny(false, jobs, timeout)
}
