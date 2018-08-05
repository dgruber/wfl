package helper

import (
	"encoding/json"
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

func ArrayJobID2GUIDs(id string) ([]string, error) {
	var guids []string
	err := json.Unmarshal([]byte(id), &guids)
	if err != nil {
		return nil, err
	}
	return guids, nil
}

func Guids2ArrayJobID(guids []string) string {
	id, err := json.Marshal(guids)
	if err != nil {
		return ""
	}
	return string(id)
}

func AddArrayJobAsSingleJobs(jt drmaa2interface.JobTemplate, t jobtracker.JobTracker, begin int, end int, step int) (string, error) {
	var guids []string
	for i := begin; i <= end; i += step {
		guid, err := t.AddJob(jt)
		if err != nil {
			return Guids2ArrayJobID(guids), err
		}
		guids = append(guids, guid)
	}
	return Guids2ArrayJobID(guids), nil
}

func IsInExpectedState(state drmaa2interface.JobState, states ...drmaa2interface.JobState) bool {
	for _, expectedState := range states {
		if state == expectedState {
			return true
		}
	}
	return false
}

func WaitForState(jt jobtracker.JobTracker, jobid string, timeout time.Duration, states ...drmaa2interface.JobState) error {
	if IsInExpectedState(jt.JobState(jobid), states...) {
		return nil
	}
	if timeout == 0 {
		return errors.New("timeout while waiting for job state")
	}

	hasStateCh := make(chan bool, 1)
	defer close(hasStateCh)

	quit := make(chan bool, 1)

	go func() {
		t := time.NewTicker(time.Millisecond * 100)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				if IsInExpectedState(jt.JobState(jobid), states...) {
					hasStateCh <- true
					return
				}
			case <-quit:
				return
			}
		}
	}()

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	select {
	case <-hasStateCh:
		return nil
	case <-ticker.C:
		quit <- true
		return errors.New("timeout while waiting for job state")
	}
	return errors.New("unreachable code in WaitForState()")
}
