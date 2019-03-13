package helper

import (
	"encoding/json"
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

// ArrayJobID2GUIDs converts the array job ID returned from
// the AddArrayJobAsSingleJobs.
func ArrayJobID2GUIDs(id string) ([]string, error) {
	var guids []string
	err := json.Unmarshal([]byte(id), &guids)
	if err != nil {
		return nil, err
	}
	return guids, nil
}

// Guids2ArrayJobID creates an array job ID out of the
// given single job IDs.
func Guids2ArrayJobID(guids []string) string {
	id, err := json.Marshal(guids)
	if err != nil {
		return ""
	}
	return string(id)
}

// AddArrayJobAsSingleJobs takes an job array definition and submits single
// jobs through the AddJob() method of the given job tracker. This function
// is typically needed when a DRM does not support job arrays natively.
// The returned array job ID is created from all of the returned job IDs and
// does not work with the DRM directly.
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

// IsInExpectedState checks if state is in one of the given states.
func IsInExpectedState(state drmaa2interface.JobState, states ...drmaa2interface.JobState) bool {
	for _, expectedState := range states {
		if state == expectedState {
			return true
		}
	}
	return false
}

// WaitForState blocks until job is in any of the given states or a timeout happens.
func WaitForState(jt jobtracker.JobTracker, jobid string, timeout time.Duration, states ...drmaa2interface.JobState) error {
	if IsInExpectedState(jt.JobState(jobid), states...) {
		return nil
	}
	if timeout == 0 {
		return errors.New("timeout while waiting for job state")
	}

	hasStateCh := make(chan bool, 1)
	defer close(hasStateCh)

	go func() {
		t := time.NewTicker(time.Millisecond * 100)
		defer t.Stop()

		timeoutTicker := time.NewTicker(timeout)
		defer timeoutTicker.Stop()

		for {
			select {
			case <-timeoutTicker.C:
				hasStateCh <- false
				return
			case <-t.C:
				if IsInExpectedState(jt.JobState(jobid), states...) {
					hasStateCh <- true
					return
				}
			}
		}
	}()

	reachedState := <-hasStateCh
	if !reachedState {
		return errors.New("timeout while waiting for job state")
	}
	return nil
}
