package simpletracker

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"strings"
	"sync"
	"time"
)

// JobTracker implements the JobTracker interface and treats
// jobs as OS processes.
type JobTracker struct {
	sync.Mutex
	jobsession string
	shutdown   bool // signal to destroy the tracker
	// communication between process trackers and registered functions for those events
	// ps stores information about state and job info of jobs
	ps *PubSub
	// stores jobs and resource usage
	js *JobStore
}

// New creates and initializes a JobTracker.
func New(jobsession string) *JobTracker {
	ps, _ := NewPubSub()
	tracker := JobTracker{
		jobsession: jobsession,
		js:         NewJobStore(),
		shutdown:   false,
		ps:         ps,
	}
	tracker.ps.StartBookKeeper()
	return &tracker
}

// Destroy signals the JobTracker to shutdown.
func (jt *JobTracker) Destroy() error {
	jt.Lock()
	defer jt.Unlock()
	jt.shutdown = true
	return nil
}

// Tracker keeps track of all jobs and updates job objects in case of changes

// ListJobs returns a list of all job IDs stored in the job store.
func (jt *JobTracker) ListJobs() ([]string, error) {
	jt.Lock()
	defer jt.Unlock()
	tmp := make([]string, len(jt.js.jobids), len(jt.js.jobids))
	copy(tmp, jt.js.jobids)
	return tmp, nil
}

// AddJob creates a process, fills in the internal job state and saves the
// job internally.
func (jt *JobTracker) AddJob(t drmaa2interface.JobTemplate) (string, error) {
	jt.Lock()
	defer jt.Unlock()

	jobid := GetNextJobID()
	jt.ps.NotifyAndWait(JobEvent{
		JobState: drmaa2interface.Queued,
		JobID:    jobid,
		JobInfo: drmaa2interface.JobInfo{
			State:          drmaa2interface.Queued,
			Slots:          1,
			SubmissionTime: time.Now(),
			ID:             jobid,
		}})

	// here also an event
	pid, err := StartProcess(jobid, 0, t, jt.ps.jobch)
	if err != nil {
		jt.ps.NotifyAndWait(JobEvent{
			JobState: drmaa2interface.Failed,
			JobID:    jobid})
		return "", err
	}
	jt.js.SaveJob(jobid, t, pid)
	return jobid, nil
}

// DeleteJob removes a job from the internal job storage but only
// when the job is in any finished state.
func (jt *JobTracker) DeleteJob(jobid string) error {
	jt.Lock()
	defer jt.Unlock()

	if !jt.js.HasJob(jobid) {
		return errors.New("Job does not exist in job store")
	}
	jt.ps.Lock()
	state, exists := jt.ps.jobState[jobid]
	jt.ps.Unlock()
	if exists && (state != drmaa2interface.Done && state != drmaa2interface.Failed) {
		return errors.New("Job is not in an end state (done/failed)")
	}
	if !exists {
		return errors.New("Job does not exist")
	}
	jt.js.RemoveJob(jobid)
	jt.ps.Unregister(jobid)
	return nil
}

func cleanup(pids []int) {
	for _, pid := range pids {
		KillPid(pid)
	}
}

// AddArrayJob starts end-begin/step processes based on the given JobTemplate.
// Note that maxParallel is not yet implemented.
func (jt *JobTracker) AddArrayJob(t drmaa2interface.JobTemplate, begin, end, step, maxParallel int) (string, error) {
	jt.Lock()
	var pids []int
	arrayjobid := GetNextJobID()
	if step <= 0 {
		step = 1
	}
	// put all jobs in queued state
	for i := begin; i <= end; i += step {
		jobid := fmt.Sprintf("%s.%d", arrayjobid, i)
		jt.ps.NotifyAndWait(JobEvent{
			JobState: drmaa2interface.Queued,
			JobID:    jobid,
			JobInfo: drmaa2interface.JobInfo{
				State:          drmaa2interface.Queued,
				Slots:          1,
				SubmissionTime: time.Now(),
				ID:             jobid,
			}})
		pids = append(pids, 0)
	}
	jt.js.SaveArrayJob(arrayjobid, pids, t, begin, end, step)
	jt.Unlock()
	errCh := arrayJobSubmissionController(jt, arrayjobid, t, begin, end, step, maxParallel)
	if err := <-errCh; err != nil {
		return "", err
	}
	return arrayjobid, nil
}

// ListArrayJobs returns all job IDs the job array ID is associated with.
func (jt *JobTracker) ListArrayJobs(id string) ([]string, error) {
	jt.Lock()
	isArray, exists := jt.js.isArrayJob[id]
	jt.Unlock()
	if !exists {
		return nil, errors.New("Array job not found")
	}
	if isArray == false {
		return nil, errors.New("Job is not an array job")
	}
	jt.Lock()
	jobids := make([]string, 0, len(jt.js.jobs[id]))
	for _, job := range jt.js.jobs[id] {
		jobids = append(jobids, fmt.Sprintf("%s.%d", id, job.TaskID))
	}
	jt.Unlock()
	return jobids, nil
}

// JobState returns the current state of the job (running, suspended, done, failed).
func (jt *JobTracker) JobState(jobid string) drmaa2interface.JobState {
	jt.Lock()
	defer jt.Unlock()
	jt.ps.Lock()
	defer jt.ps.Unlock()

	// job state:
	// ----------

	// Triggered:
	//
	// AddJob --> Running or Failed
	// DeleteJob --> removes job when it is in end state
	// JobControl --> Suspended / Running

	// Async:
	//
	// watch() --> (pubsub) StartBookKeeper() -> StartProcess() --> Done / Failed // ==> in PubSub

	state, exists := jt.ps.jobState[jobid]
	if !exists {
		state = drmaa2interface.Undetermined
	}
	return state
}

// JobInfo returns more detailed information about a job.
func (jt *JobTracker) JobInfo(jobid string) (drmaa2interface.JobInfo, error) {
	jt.Lock()
	defer jt.Unlock()
	jt.ps.Lock()
	ji, exists := jt.ps.jobInfo[jobid]
	jt.ps.Unlock()
	if exists == true {
		return ji, nil
	}
	return ji, errors.New("job does not exist")
}

// JobControl suspends, resumes, or terminates a job.
func (jt *JobTracker) JobControl(jobid, state string) error {
	jt.Lock()
	defer jt.Unlock()

	pid, err := jt.js.GetPID(jobid)
	if err != nil {
		return errors.New("job does not exist")
	}

	switch state {
	case "suspend":
		if pid == 0 {
			return errors.New("job is not running")
		}
		err := SuspendPid(pid)
		if err == nil {
			jt.ps.Lock()
			jt.ps.jobState[jobid] = drmaa2interface.Suspended
			jt.ps.Unlock()
		}
		return err
	case "resume":
		if pid == 0 {
			return errors.New("job is not running")
		}
		err := ResumePid(pid)
		if err == nil {
			jt.ps.Lock()
			jt.ps.jobState[jobid] = drmaa2interface.Running
			jt.ps.Unlock()
		}
		return err
	case "hold":
		return errors.New("Unsupported Operation")
	case "release":
		return errors.New("Unsupported Operation")
	case "terminate":
		// if job is queued - terminate it by setting it to
		// failed state (see arrayJobSubmissionController())
		jt.ps.Lock()
		state := jt.ps.jobState[jobid]
		if state == drmaa2interface.Queued {
			jt.ps.jobState[jobid] = drmaa2interface.Failed
			jt.ps.Unlock()
			return nil
		}
		jt.ps.Unlock()
		err := KillPid(pid)
		if err == nil {
			jt.ps.Lock()
			jt.ps.jobState[jobid] = drmaa2interface.Failed
			jt.ps.Unlock()
		}
		return err
	}

	return errors.New("undefined state")
}

// Wait blocks until the job with the given job id is in on of the given states.
// If the job is after the given duration is still not in any of the states the
// method returns an error. If the duration is 0 then it waits infitely.
func (jt *JobTracker) Wait(jobid string, d time.Duration, state ...drmaa2interface.JobState) error {
	var timeoutCh <-chan time.Time
	if d.Seconds() == 0.0 {
		// infinite
		timeoutCh = make(chan time.Time)
	} else {
		// create timeout channel
		timeoutCh = time.Tick(d)
	}

	// jobid can be a job or array job task
	jobparts := strings.Split(jobid, ".")

	// check if job exists and if it is in an end state already which does not change
	jt.Lock()
	_, exists := jt.js.jobs[jobparts[0]]
	if exists == false {
		jt.Unlock()
		return errors.New("job does not exist")
	}

	// register channel to get informed when job finished or reached the state
	waitChannel, err := jt.ps.Register(jobid, state...)
	jt.Unlock()
	if err != nil {
		return err
	}
	if waitChannel == nil {
		// we are already in expected state
		return nil
	}

	select {
	case newState := <-waitChannel:
		// end states are reported as well
		for i := range state {
			if newState == state[i] {
				return nil
			}
		}
		return drmaa2interface.Error{Message: "Job finished with different state", ID: drmaa2interface.Internal}
	case <-timeoutCh:
		return drmaa2interface.Error{Message: "Timeout occurred while waiting for job state", ID: drmaa2interface.Timeout}
	}
}

// ListJobCategories returns an empty list as JobCategories are
// currently not defined for OS processes.
func (jt *JobTracker) ListJobCategories() ([]string, error) {
	return []string{}, nil
}
