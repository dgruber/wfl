package drmaa2interface

import (
	"time"
)

// JobInfo represents the state of a job.
type JobInfo struct {
	Extensible
	Extension         `xml:"-" json:"-"`
	ID                string        `json:"id"`
	ExitStatus        int           `json:"exitStatus"`
	TerminatingSignal string        `json:"terminationSignal"`
	Annotation        string        `json:"annotation"`
	State             JobState      `json:"state"`
	SubState          string        `json:"subState"`
	AllocatedMachines []string      `json:"allocatedMachines"`
	SubmissionMachine string        `json:"submissionMachine"`
	JobOwner          string        `json:"jobOwner"`
	Slots             int64         `json:"slots"`
	QueueName         string        `json:"queueName"`
	WallclockTime     time.Duration `json:"wallockTime"`
	CPUTime           int64         `json:"cpuTime"`
	SubmissionTime    time.Time     `json:"submissionTime"`
	DispatchTime      time.Time     `json:"dispatchTime"`
	FinishTime        time.Time     `json:"finishTime"`
}

// CreateJobInfo creates a JobInfo object where all values are initialized
// with UNSET (needed in order to differentiate if a value is
// not set or 0).
func CreateJobInfo() (ji JobInfo) {
	ji.ExitStatus = UnsetNum
	ji.Slots = UnsetNum
	ji.CPUTime = UnsetTime
	ji.State = Unset
	return ji
}
