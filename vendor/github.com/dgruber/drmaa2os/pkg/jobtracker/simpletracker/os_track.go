package simpletracker

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// TrackProcess supervises a running process and sends a notification when
// the process is finished.
func TrackProcess(cmd *exec.Cmd, jobid string, startTime time.Time,
	finishedJobChannel chan JobEvent, waitForFiles int, waitCh chan bool) {
	state, err := cmd.Process.Wait()

	// wait until all filedescriptors (stdout, stderr) of the
	// process are closed
	for waitForFiles > 0 {
		<-waitCh
		waitForFiles--
	}

	if err != nil {
		ji := makeLocalJobInfo()
		ji.State = drmaa2interface.Failed
		finishedJobChannel <- JobEvent{
			JobState: drmaa2interface.Failed,
			JobID:    jobid,
			JobInfo:  ji,
		}
		return
	}

	ji := collectUsage(state, jobid, startTime)
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

func collectUsage(state *os.ProcessState, jobid string, startTime time.Time) drmaa2interface.JobInfo {
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

	ji.WallclockTime = time.Since(startTime)
	ji.CPUTime = 0
	ji.ID = jobid
	ji.QueueName = ""

	return ji
}
