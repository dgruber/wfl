package simpletracker

import (
	_ "github.com/dgruber/drmaa2interface"
	"time"
)

func watch(tracker *JobTracker) {
	// keep jobs state up to date
	tracker.ps.StartBookKeeper()

	// performs frequent checks of all jobs and update jobstate
	t := time.NewTicker(time.Second)
	for range t.C {
		tracker.Lock()
		if tracker.shutdown == true {
			return
		}
		// update job state???
		tracker.Unlock()
	}
}
