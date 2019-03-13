package d2hlp

import (
	"github.com/dgruber/drmaa2interface"
	"time"
)

func JobInfoMatches(ji drmaa2interface.JobInfo, filter drmaa2interface.JobInfo) bool {
	if filter.ID != "" {
		if ji.ID != filter.ID {
			return false
		}
	}
	if filter.ExitStatus != drmaa2interface.UnsetNum {
		if ji.ExitStatus != filter.ExitStatus {
			return false
		}
	}
	if filter.TerminatingSignal != "" {
		if ji.TerminatingSignal != filter.TerminatingSignal {
			return false
		}
	}
	if filter.Annotation != "" {
		if ji.Annotation != filter.Annotation {
			return false
		}
	}
	if filter.State != drmaa2interface.Unset {
		if ji.State != filter.State {
			return false
		}
	}
	if filter.SubState != "" {
		if ji.SubState != filter.SubState {
			return false
		}
	}
	if filter.AllocatedMachines != nil {
		// must run on a superset of the given machines
		if len(ji.AllocatedMachines) < len(filter.AllocatedMachines) {
			return false
		}

		for _, v := range filter.AllocatedMachines {
			found := false
			for _, i := range ji.AllocatedMachines {
				if v == i {
					found = true
					break
				}
			}
			if found == false {
				return false
			}
		}
	}
	if filter.SubmissionMachine != "" {
		if ji.SubmissionMachine != filter.SubmissionMachine {
			return false
		}
	}
	if filter.JobOwner != "" {
		if ji.JobOwner != filter.JobOwner {
			return false
		}
	}
	if filter.Slots != drmaa2interface.UnsetNum {
		if ji.Slots != filter.Slots {
			return false
		}
	}
	if filter.QueueName != "" {
		if ji.QueueName != filter.QueueName {
			return false
		}
	}
	if filter.WallclockTime != 0 {
		if ji.WallclockTime < filter.WallclockTime {
			return false
		}
	}
	if filter.CPUTime != drmaa2interface.UnsetTime {
		if ji.CPUTime < filter.CPUTime {
			return false
		}
	}
	var nullTime time.Time
	if filter.SubmissionTime != nullTime {
		if ji.SubmissionTime.Before(filter.SubmissionTime) {
			return false
		}
	}
	if filter.DispatchTime != nullTime {
		if ji.DispatchTime.Before(filter.DispatchTime) {
			return false
		}
	}
	if filter.FinishTime != nullTime {
		if ji.FinishTime.Before(filter.FinishTime) {
			return false
		}
	}
	return true
}
