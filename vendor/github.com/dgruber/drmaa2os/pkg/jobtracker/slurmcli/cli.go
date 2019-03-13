package slurmcli

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"os/exec"
)

// Slurm is a wrapper for the slurm CLI tools.
type Slurm struct {
	queue           string
	batch           string
	control         string
	cancel          string
	acct            string
	suspendBySignal bool
}

// NewSlurm creates a new wrapper for slurm CLI tools.
// It uses the given command line tools for the calls.
func NewSlurm(sbatch, squeue, scontrol, scancel, sacct string, suspendBySignal bool) *Slurm {
	return &Slurm{
		batch:           sbatch,
		queue:           squeue,
		control:         scontrol,
		cancel:          scancel,
		acct:            sacct,
		suspendBySignal: suspendBySignal,
	}
}

// ListJobs returns all jobs for a given account in a given state.
func (s *Slurm) ListJobs(account, states string) ([]string, error) {
	out, err := run(s.queue, "-h", "-A", account, "--states="+states)
	if err != nil {
		return nil, err
	}
	return parsesqueue(out)
}

// SubmitJob converts the job template into job submission options
// and submits a job with sbatch.
func (s *Slurm) SubmitJob(account string, jt drmaa2interface.JobTemplate) (string, error) {
	args, err := convertJobTemplate(account, jt)
	if err != nil {
		return "", err
	}
	out, err := run(s.batch, args...)
	if err != nil {
		return "", err
	}
	return parsesbatch(out)
}

// SubmitJobArray converts a job template into job submission
// options and submits an job array (like sbatch --array=1-7:2).
func (s *Slurm) SubmitJobArray(account string, jt drmaa2interface.JobTemplate, start, end, step, maxParallel int) (string, error) {
	args, err := convertJobTemplate(account, jt)
	if err != nil {
		return "", err
	}
	arrayJobArgs, err := convertArrayJobArgs(start, end, step, maxParallel)
	if err != nil {
		return "", err
	}
	args = append(arrayJobArgs, args...)
	out, err := run(s.batch, args...)
	if err != nil {
		return "", err
	}
	return parsesbatch(out)
}

// Suspend sends either a SIGTSTP signal to a job or releases the jobs
// resources (admin rights required).
func (s *Slurm) Suspend(account, jobid string) error {
	var err error
	if s.suspendBySignal {
		_, err = run(s.cancel, "--signal=SIGSTP", jobid)
	} else {
		_, err = run(s.control, "suspend", jobid)
	}
	return err
}

// Resume sends either a SIGCONT signal to a job or re-claims the jobs
// resources (admin rights required).
func (s *Slurm) Resume(account, jobid string) error {
	var err error
	if s.suspendBySignal {
		_, err = run(s.cancel, "--signal=SIGCONT", jobid)
	} else {
		_, err = run(s.control, "resume", jobid)
	}
	return err
}

// Terminate stops a job from execution.
func (s *Slurm) Terminate(account, jobid string) error {
	_, err := run(s.cancel, "-A", account, jobid)
	return err
}

// State return the state of a given job.
func (s *Slurm) State(account, jobid string) drmaa2interface.JobState {
	// sacct -A default -j 25.batch --parsable2 -o "State" -n
	// RUNNING
	state, err := run(s.acct, "-A", account, "-j",
		jobid+".batch", "--parsable2", "-o", "\"State\"", "-n")
	if err != nil {
		return drmaa2interface.Undetermined
	}
	return convertState(string(state))
}

func run(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("(%s %v) command failed with %s", command, args, err.Error())
	}
	if !cmd.ProcessState.Success() {
		return nil, fmt.Errorf("(%s %v) command failed", command, args)
	}
	return out, err
}

// CheckCLI tests of all commmand line applications can be called.
func CheckCLI(slurm *Slurm) error {
	if !cmdExists(slurm.batch) {
		return fmt.Errorf("sbatch command (%s) does not exist", slurm.batch)
	}
	if !cmdExists(slurm.queue) {
		return fmt.Errorf("squeue command (%s) does not exist", slurm.queue)
	}
	if !cmdExists(slurm.control) {
		return fmt.Errorf("scontrol command (%s) does not exist", slurm.control)
	}
	if !cmdExists(slurm.cancel) {
		return fmt.Errorf("scancel command (%s) does not exist", slurm.cancel)
	}
	if !cmdExists(slurm.acct) {
		return fmt.Errorf("sacct command (%s) does not exist", slurm.acct)
	}
	return nil
}

func cmdExists(cli string) bool {
	cmd := exec.Command(cli, "--help")
	err := cmd.Run()
	if err != nil || !cmd.ProcessState.Success() {
		return false
	}
	return true
}
