package simpletracker

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"strconv"
	"strings"
)

type JobStore struct {
	// jobids contains all known jobs in the system until they are reaped (Reap())
	// these are jobs, not array jobs and can be in format "1.1" or "1"
	jobids []string
	// running jobs
	// string is jobid and isArrayJob determines the type
	templates  map[string]drmaa2interface.JobTemplate
	jobs       map[string][]InternalJob
	isArrayJob map[string]bool
}

func NewJobStore() *JobStore {
	return &JobStore{
		jobids:     make([]string, 0, 512),
		templates:  make(map[string]drmaa2interface.JobTemplate),
		jobs:       make(map[string][]InternalJob),
		isArrayJob: make(map[string]bool),
	}
}

func (js *JobStore) SaveJob(jobid string, t drmaa2interface.JobTemplate, pid int) {
	js.templates[jobid] = t
	js.jobids = append(js.jobids, jobid)
	js.jobs[jobid] = []InternalJob{
		InternalJob{State: drmaa2interface.Running, PID: pid},
	}
}

func (js *JobStore) SaveArrayJob(arrayjobid string, pids []int, t drmaa2interface.JobTemplate, begin int, end int, step int) {
	pid := 0
	js.templates[arrayjobid] = t
	js.isArrayJob[arrayjobid] = true
	js.jobs[arrayjobid] = make([]InternalJob, 0, (end-begin)/step)

	for i := begin; i <= end; i += step {
		jobid := fmt.Sprintf("%s.%d", arrayjobid, i)
		js.jobids = append(js.jobids, jobid)
		js.jobs[arrayjobid] = append(js.jobs[arrayjobid], InternalJob{TaskID: i, State: drmaa2interface.Running, PID: pids[pid]})
		pid++
	}
}

func (js *JobStore) GetPID(jobid string) (int, error) {
	jobelements := strings.Split(jobid, ".")
	if job, exists := js.jobs[jobelements[0]]; !exists {
		return -1, errors.New("Job does not exist")
	} else {
		var (
			taskid int
			err    error
		)
		if len(jobelements) > 1 {
			// is array job
			taskid, err = strconv.Atoi(jobelements[1])
			if err != nil {
				return -1, errors.New("TaskID within job ID is not a number")
			}
		}
		if taskid == 0 || taskid == 1 {
			return job[0].PID, nil
		}
		for task, _ := range job {
			if job[task].TaskID == taskid {
				return job[task].PID, nil
			}
		}
	}
	return -1, errors.New("TaskID not found in job array")
}
