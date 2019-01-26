package drmaa2os

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/d2hlp"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

// JobSession instance acts as container for job instances controlled
// through the DRMAA API. The session methods support the submission
// of new jobs and the monitoring of existing jobs.
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
	var lastError error
	jobCategories := make([]string, 0, 16)
	for _, tracker := range js.tracker {
		cat, err := tracker.ListJobCategories()
		if err != nil {
			lastError = err
			continue
		}
		jobCategories = append(jobCategories, cat...)
	}
	return jobCategories, lastError
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

					job := newJob(jobid, js.name, jobtemplate, tracker)
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

		job := newJob(id, js.name, jobtemplate, js.tracker[0])
		joblist = append(joblist, drmaa2interface.Job(job))
	}
	return NewArrayJob(id, js.name, drmaa2interface.JobTemplate{}, joblist), nil
}

// RunJob method submits a job with the attributes defined in the given job template
// instance. The method returns a Job object that represents the job in the underlying
// DRM system. Depending on the job template settings, submission attempts may be
// rejected with an InvalidArgumentException. The error details SHOULD provide further
// information about the attribute(s) responsible for the rejection. When this method
// returns a valid Job instance, the following conditions SHOULD be fulfilled:
// - The job is part of the persistent state of the job session.
// - All non-DRMAA and DRMAA interfaces to the DRM system report the job as
//   being submitted to the DRM system.
// - The job has one of the DRMAA job states.
func (js *JobSession) RunJob(jt drmaa2interface.JobTemplate) (drmaa2interface.Job, error) {
	id, err := js.tracker[0].AddJob(jt)
	if err != nil {
		return nil, err
	}
	return newJob(id, js.name, jt, js.tracker[0]), nil
}

// RunBulkJobs method creates a set of parametric jobs, each with attributes as defined
// in the given job template instance.
func (js *JobSession) RunBulkJobs(jt drmaa2interface.JobTemplate, begin, end, step, maxParallel int) (drmaa2interface.ArrayJob, error) {
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
				abort <- true
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
