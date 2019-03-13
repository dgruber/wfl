package simpletracker

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"sync"
)

// JobEvent is send whenever a job status change is happening
// to inform all registered listeners.
type JobEvent struct {
	JobID    string
	JobState drmaa2interface.JobState
	JobInfo  drmaa2interface.JobInfo
	callback chan bool // if set sends true if event was distributed
}

// PubSub distributes job status change events to clients which
// Register() at PubSub.
type PubSub struct {
	sync.Mutex

	// go routines write into that channel when process has finished
	jobch chan JobEvent

	// maps a jobid to functions registered for waiting for a specific
	// state of that job
	waitFunctions map[string][]waitRequest

	// feed by bookKeeper: current state
	jobState map[string]drmaa2interface.JobState
	jobInfo  map[string]drmaa2interface.JobInfo
}

// NewPubSub returns an initialized PubSub structure and
// the JobEvent channel which is used by the caller to publish
// job events (i.e. job state transitions).
func NewPubSub() (*PubSub, chan JobEvent) {
	jeCh := make(chan JobEvent, 1)
	return &PubSub{
		jobch:         jeCh,
		waitFunctions: make(map[string][]waitRequest),
		jobState:      make(map[string]drmaa2interface.JobState),
		jobInfo:       make(map[string]drmaa2interface.JobInfo),
	}, jeCh
}

// Register returns a channel which emits a job state once the given
// job transitions in one of the given states. If job is already
// in the expected state it returns nil as channel and nil as error.
//
// TODO add method for removing specific wait functions.
func (ps *PubSub) Register(jobid string, states ...drmaa2interface.JobState) (chan drmaa2interface.JobState, error) {
	ps.Lock()
	defer ps.Unlock()

	// check if job is already in the expected state
	state, exists := ps.jobState[jobid]
	if exists {
		for _, expectedState := range states {
			if expectedState == state {
				return nil, nil
			}
		}
		if state == drmaa2interface.Failed || state == drmaa2interface.Done {
			return nil, errors.New("job already finished")
		}
	}

	waitChannel := make(chan drmaa2interface.JobState, 1)
	ps.waitFunctions[jobid] = append(ps.waitFunctions[jobid],
		waitRequest{ExpectedState: states, WaitChannel: waitChannel})
	return waitChannel, nil
}

// Unregister removes all functions waiting for a specific job and
// all occurences of the job itself.
func (ps *PubSub) Unregister(jobid string) {
	ps.Lock()
	defer ps.Unlock()
	delete(ps.waitFunctions, jobid)
	delete(ps.jobState, jobid)
	delete(ps.jobInfo, jobid)
}

// NotifyAndWait sends a job event and waits until it was distributed
// to all waiting functions.
func (ps *PubSub) NotifyAndWait(evt JobEvent) {
	evt.callback = make(chan bool, 1)
	ps.jobch <- evt
	<-evt.callback
}

// waitRequest defines when to notify (ExpectedState) and where to notify (WaitChannel)
type waitRequest struct {
	ExpectedState []drmaa2interface.JobState
	WaitChannel   chan drmaa2interface.JobState
}

// StartBookKeeper processes all job state changes from the process trackers
// and notifies registered wait functions.
func (ps *PubSub) StartBookKeeper() {
	go func() {
		for event := range ps.jobch {
			ps.Lock()
			// inform registered functions
			for _, waiter := range ps.waitFunctions[event.JobID] {
				// inform when expected state is reached
				for i := range waiter.ExpectedState {
					if event.JobState == waiter.ExpectedState[i] {
						waiter.WaitChannel <- event.JobState
					}
				}
			}
			ps.jobState[event.JobID] = event.JobState
			if info, exists := ps.jobInfo[event.JobID]; exists {
				ps.jobInfo[event.JobID] = mergeJobInfo(info, event.JobInfo)
			} else {
				// TODO deep copy
				ps.jobInfo[event.JobID] = event.JobInfo
			}

			ps.Unlock()
			if event.callback != nil {
				event.callback <- true
			}
		}
	}()
}

func mergeJobInfo(oldJI, newJI drmaa2interface.JobInfo) drmaa2interface.JobInfo {
	if newJI.ID != "" {
		oldJI.ID = newJI.ID
	}
	if newJI.ExitStatus != 0 {
		oldJI.ExitStatus = newJI.ExitStatus
	}
	if newJI.TerminatingSignal != "" {
		oldJI.TerminatingSignal = newJI.TerminatingSignal
	}
	if newJI.Annotation != "" {
		oldJI.Annotation = newJI.Annotation
	}
	if newJI.State != drmaa2interface.Unset {
		oldJI.State = newJI.State
	}
	if newJI.SubState != "" {
		oldJI.SubState = newJI.SubState
	}
	if newJI.AllocatedMachines != nil {
		oldJI.AllocatedMachines = make([]string, 0, len(newJI.AllocatedMachines))
		copy(oldJI.AllocatedMachines, newJI.AllocatedMachines)
	}
	if newJI.SubmissionMachine != "" {
		oldJI.SubmissionMachine = newJI.SubmissionMachine
	}
	if newJI.JobOwner != "" {
		oldJI.JobOwner = newJI.JobOwner
	}
	if newJI.Slots != 0 {
		oldJI.Slots = newJI.Slots
	}
	if newJI.QueueName != "" {
		oldJI.QueueName = newJI.QueueName
	}
	if newJI.WallclockTime != 0.0 {
		oldJI.WallclockTime = newJI.WallclockTime
	}
	if newJI.CPUTime != 0.0 {
		oldJI.CPUTime = newJI.CPUTime
	}
	if !newJI.SubmissionTime.IsZero() {
		oldJI.SubmissionTime = newJI.SubmissionTime
	}
	if !newJI.DispatchTime.IsZero() {
		oldJI.DispatchTime = newJI.DispatchTime
	}
	if !newJI.FinishTime.IsZero() {
		oldJI.FinishTime = newJI.FinishTime
	}
	return oldJI
}
