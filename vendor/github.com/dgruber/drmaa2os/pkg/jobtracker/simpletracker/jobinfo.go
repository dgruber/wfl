package simpletracker

import (
	"github.com/dgruber/drmaa2interface"
)

// JobInfo requires PubSub as it forwards and saves state (it has the jobs
// connected through os StartProcess() goroutine)

func CreateJobInfo() drmaa2interface.JobInfo {
	ji := drmaa2interface.JobInfo{}

	return ji
}
