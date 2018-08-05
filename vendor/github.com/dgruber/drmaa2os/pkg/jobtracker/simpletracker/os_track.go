package simpletracker

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func TrackProcess(cmd *exec.Cmd, jobid string, finishedJobChannel chan JobEvent, waitForFiles int, waitCh chan bool) {
	// supervise process

	dispatchTime := time.Now()

	state, err := cmd.Process.Wait()

	for waitForFiles > 0 {
		<-waitCh
		waitForFiles--
	}

	if err != nil {
		ji := makeLocalJobInfo()
		ji.State = drmaa2interface.Undetermined
		finishedJobChannel <- JobEvent{JobState: drmaa2interface.Failed, JobID: jobid, JobInfo: ji}
		return
	}

	ji := collectUsage(state, jobid, dispatchTime)
	finishedJobChannel <- JobEvent{JobState: ji.State, JobID: jobid, JobInfo: ji}
}

func makeLocalJobInfo() drmaa2interface.JobInfo {
	host, _ := os.Hostname()

	return drmaa2interface.JobInfo{
		AllocatedMachines: []string{host},
		FinishTime:        time.Now(),
		SubmissionMachine: host,
		JobOwner:          fmt.Sprintf("%d", os.Getuid()),
	}
}

func collectUsage(state *os.ProcessState, jobid string, dispatchTime time.Time) drmaa2interface.JobInfo {
	ji := makeLocalJobInfo()
	ji.State = drmaa2interface.Undetermined

	if status, ok := state.Sys().(syscall.WaitStatus); ok {
		ji.ExitStatus = status.ExitStatus()
		ji.TerminatingSignal = status.Signal().String()
	}

	if usage, ok := state.SysUsage().(syscall.Rusage); ok {
		ji.CPUTime = usage.Utime.Sec + usage.Stime.Sec
		// TODO extensions
	}

	if state != nil && state.Success() {
		ji.State = drmaa2interface.Done
	} else {
		ji.State = drmaa2interface.Failed
	}

	if ji.ExitStatus != 0 {
		ji.State = drmaa2interface.Failed
	}

	ji.WallclockTime = time.Since(dispatchTime)
	ji.CPUTime = 0
	ji.DispatchTime = dispatchTime
	ji.ID = jobid
	ji.QueueName = ""
	ji.Slots = 1
	ji.SubmissionTime = dispatchTime

	return ji
}
