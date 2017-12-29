package simpletracker

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"sync"
)

type JobEvent struct {
	JobID    string
	JobState drmaa2interface.JobState
	JobInfo  drmaa2interface.JobInfo
}

type PubSub struct {
	sync.Mutex

	// go routines write into that channel when process has finished
	jobch chan JobEvent

	// maps a jobid to functions registered for waiting for a specific state of that job
	waitFunctions map[string][]waitRequest

	// feed by bookKeeper: current state
	jobState        map[string]drmaa2interface.JobState
	jobInfoFinished map[string]drmaa2interface.JobInfo
}

func NewPubSub() (*PubSub, chan JobEvent) {
	jeCh := make(chan JobEvent, 1)
	return &PubSub{
		jobch:           jeCh,
		waitFunctions:   make(map[string][]waitRequest),
		jobState:        make(map[string]drmaa2interface.JobState),
		jobInfoFinished: make(map[string]drmaa2interface.JobInfo),
	}, jeCh
}

func (ps *PubSub) Register(jobid string, states ...drmaa2interface.JobState) (chan drmaa2interface.JobState, error) {
	ps.Lock()
	defer ps.Unlock()

	// check if job already finished
	if state, exists := ps.jobState[jobid]; exists {
		if state == drmaa2interface.Failed || state == drmaa2interface.Done {
			return nil, errors.New("job already finished")
		}
	}

	waitChannel := make(chan drmaa2interface.JobState, 1)
	ps.waitFunctions[jobid] = append(ps.waitFunctions[jobid], waitRequest{ExpectedState: states, WaitChannel: waitChannel})
	return waitChannel, nil
}

func (ps *PubSub) Unregister(id string) {
	// TODO
}

// waitRequest defines when to notify (ExpectedState) and where to notify (WaitChann)
type waitRequest struct {
	ExpectedState []drmaa2interface.JobState
	WaitChannel   chan drmaa2interface.JobState
}

// BookKeeper processes all job state changes from the process trackers
// and notifies registered wait functions
func (ps *PubSub) StartBookKeeper() {
	go func() {
		for event := range ps.jobch {
			ps.Lock()
			// inform registered functions
			for _, waiter := range ps.waitFunctions[event.JobID] {
				for i := range waiter.ExpectedState {
					if event.JobState == waiter.ExpectedState[i] {
						waiter.WaitChannel <- event.JobState
						continue
					}
				}
			}
			// make the job state persistent
			ps.jobState[event.JobID] = event.JobState
			// make job info persistent
			ps.jobInfoFinished[event.JobID] = event.JobInfo
			ps.Unlock()
		}
	}()
}
