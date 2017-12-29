package simpletracker

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"os"
	"strings"
	"sync"
	"time"
)

type JobTracker struct {
	sync.Mutex
	jobsession string

	// Destroy the tracker
	shutdown bool
	// communication between process trackers and registered functions for those events
	// ps stores information about state and job info of jobs
	ps *PubSub

	js *JobStore
}

func New(jobsession string) *JobTracker {
	ps, _ := NewPubSub()
	tracker := JobTracker{
		jobsession: jobsession,
		js:         NewJobStore(),
		shutdown:   false,
		ps:         ps,
	}
	go watch(&tracker)
	return &tracker
}

func (jt *JobTracker) Destroy() error {
	jt.Lock()
	defer jt.Unlock()
	jt.shutdown = true
	return nil
}

// Tracker keeps track of all jobs and updates job objects in case of changes

func (jt *JobTracker) ListJobs() ([]string, error) {
	jt.Lock()
	defer jt.Unlock()
	tmp := make([]string, len(jt.js.jobids), len(jt.js.jobids))
	copy(tmp, jt.js.jobids)
	return tmp, nil
}

func (jt *JobTracker) AddJob(t drmaa2interface.JobTemplate) (string, error) {
	jt.Lock()
	defer jt.Unlock()
	jt.ps.Lock()
	defer jt.ps.Unlock()
	jobid := GetNextJobID()

	if pid, err := StartProcess(jobid, t, jt.ps.jobch); err != nil {
		jt.ps.jobState[jobid] = drmaa2interface.Failed
		return "", err
	} else {
		jt.ps.jobState[jobid] = drmaa2interface.Running
		jt.js.SaveJob(jobid, t, pid)
	}
	return jobid, nil
}

// TODO TEST IMPLEMENTATION
func (jt *JobTracker) DeleteJob(jobid string) error {
	jt.Lock()
	defer jt.Unlock()
	if state, exists := jt.ps.jobState[jobid]; exists != true {
		return errors.New("Job does not exist")
	} else {
		if state != drmaa2interface.Done && state != drmaa2interface.Failed {
			return errors.New("Job is not in an end state (done/failed)")
		}
		// TODO delete entry
	}
	return nil
}

func cleanup(pids []int) {
	for _, pid := range pids {
		KillPid(pid)
	}
}

func (jt *JobTracker) AddArrayJob(t drmaa2interface.JobTemplate, begin int, end int, step int, maxParallel int) (string, error) {
	arrayjobid := GetNextJobID()

	// maxParallel has no meaning yet - start all processes
	var pids []int
	for i := begin; i <= end; i += step {
		jobid := fmt.Sprintf("%s.%d", arrayjobid, i)
		if pid, err := StartProcess(jobid, t, jt.ps.jobch); err != nil {
			cleanup(pids)
			return "", err
		} else {
			pids = append(pids, pid)
		}
	}

	jt.Lock()
	defer jt.Unlock()

	jt.js.SaveArrayJob(arrayjobid, pids, t, begin, end, step)

	return arrayjobid, nil
}

func (jt *JobTracker) ListArrayJobs(id string) ([]string, error) {
	if isArray, exists := jt.js.isArrayJob[id]; !exists {
		return nil, errors.New("Array job not found")
	} else {
		if isArray == false {
			return nil, errors.New("Job is not an array job")
		}
	}
	jobids := make([]string, 0, len(jt.js.jobs[id]))
	for _, job := range jt.js.jobs[id] {
		jobids = append(jobids, fmt.Sprintf("%s.%d", id, job.TaskID))
	}
	return jobids, nil
}

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

	return jt.ps.jobState[jobid]
}

func (jt *JobTracker) ProcessToJobInfo(jobid string, pid int) (drmaa2interface.JobInfo, error) {
	host, _ := os.Hostname()
	return drmaa2interface.JobInfo{
		Slots:             1,
		ID:                jobid,
		SubmissionMachine: host,
		State:             drmaa2interface.Running,
		JobOwner:          fmt.Sprintf("%d", os.Getuid()),
	}, nil
}

func (jt *JobTracker) JobInfo(jobid string) (drmaa2interface.JobInfo, error) {
	jt.Lock()
	defer jt.Unlock()

	// check finished jobs
	jt.ps.Lock()
	ji, exists := jt.ps.jobInfoFinished[jobid]
	jt.ps.Unlock()
	if exists == true {
		return ji, nil
	}

	if pid, err := jt.js.GetPID(jobid); err != nil {
		return drmaa2interface.JobInfo{}, err
	} else {
		return jt.ProcessToJobInfo(jobid, pid)
	}
}

func (jt *JobTracker) JobControl(jobid, state string) error {
	jt.Lock()
	defer jt.Unlock()

	pid, err := jt.js.GetPID(jobid)
	if err != nil {
		return errors.New("job does not exist")
	}

	switch state {
	case "suspend":
		err := SuspendPid(pid)
		if err == nil {
			jt.ps.Lock()
			jt.ps.jobState[jobid] = drmaa2interface.Suspended
			jt.ps.Unlock()
		}
		return err
	case "resume":
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
	jobidOrArrayJobId := jobparts[0]

	// check if job exists and if it is in an end state already which does not change
	jt.Lock()
	if _, exists := jt.js.jobs[jobidOrArrayJobId]; exists == false {
		jt.Unlock()
		return errors.New("job does not exist")
	} else {
		jt.ps.Lock()
		// works with jobid???
		if js, jsexists := jt.ps.jobState[jobid]; jsexists {
			if js == drmaa2interface.Failed || js == drmaa2interface.Done {
				jt.ps.Unlock()
				jt.Unlock()
				for i := range state {
					if state[i] == js {
						return nil
					}
				}
				// TODO drmaa2 error?
				return errors.New("Invalid state")
			}
		}
		jt.ps.Unlock()
	}

	// register channel to get informed when job finished or reached the state
	waitChannel, err := jt.ps.Register(jobid, state...)
	jt.Unlock()
	if err != nil {
		return err
	}

	select {
	case newState := <-waitChannel:
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

func (jt *JobTracker) ListJobCategories() ([]string, error) {
	return []string{}, nil
}
