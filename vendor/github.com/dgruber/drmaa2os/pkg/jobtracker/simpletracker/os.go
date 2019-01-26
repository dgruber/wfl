package simpletracker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/dgruber/drmaa2interface"
)

func currentEnv() map[string]string {
	env := make(map[string]string, len(os.Environ()))
	for _, e := range os.Environ() {
		env[e] = os.Getenv(e)
	}
	return env
}

func restoreEnv(env map[string]string) {
	for _, e := range os.Environ() {
		os.Unsetenv(e)
	}
	for key, value := range env {
		os.Setenv(key, value)
	}
}

// StartProcess creates a new process based on the JobTemplate.
// It returns the PID or 0 and an error if the process could be
// created. The given channel is used for communicating back
// when the job state changed.
func StartProcess(jobid string, task int, t drmaa2interface.JobTemplate, finishedJobChannel chan JobEvent) (int, error) {
	cmd := exec.Command(t.RemoteCommand, t.Args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if valid, err := validateJobTemplate(t); valid == false {
		return 0, err
	}

	waitForFiles := 0
	waitCh := make(chan bool, 3)

	if t.InputPath != "" {
		if stdin, err := cmd.StdinPipe(); err == nil {
			waitForFiles++
			redirectIn(stdin, t.InputPath, waitCh)
		}
	}
	if t.OutputPath != "" {
		if stdout, err := cmd.StdoutPipe(); err == nil {
			waitForFiles++
			redirectOut(stdout, t.OutputPath, waitCh)
		}
	}
	if t.ErrorPath != "" {
		if stderr, err := cmd.StderrPipe(); err == nil {
			waitForFiles++
			redirectOut(stderr, t.ErrorPath, waitCh)
		}
	}

	var mtx sync.Mutex

	mtx.Lock()
	env := currentEnv()

	for key, value := range t.JobEnvironment {
		os.Setenv(key, value)
	}
	os.Setenv("JOB_ID", jobid)
	if task != 0 {
		os.Setenv("TASK_ID", fmt.Sprintf("%d", task))
	}

	if err := cmd.Start(); err != nil {
		mtx.Unlock()
		return 0, err
	}

	host, _ := os.Hostname()
	startTime := time.Now()

	finishedJobChannel <- JobEvent{
		JobState: drmaa2interface.Running,
		JobID:    jobid,
		JobInfo: drmaa2interface.JobInfo{
			State:             drmaa2interface.Running,
			DispatchTime:      startTime,
			AllocatedMachines: []string{host},
		},
	}

	go TrackProcess(cmd, jobid, startTime, finishedJobChannel, waitForFiles, waitCh)

	restoreEnv(env)
	mtx.Unlock()

	if cmd.Process == nil {
		return 0, errors.New("process is nil")
	}
	return cmd.Process.Pid, nil
}

func redirectOut(src io.ReadCloser, outfilename string, waitCh chan bool) {
	go func() {
		buf := make([]byte, 1024)
		outfile, _ := os.Create(outfilename)
		io.CopyBuffer(outfile, src, buf)
		outfile.Close()
		waitCh <- true
	}()
}

func redirectIn(out io.WriteCloser, infilename string, waitCh chan bool) {
	go func() {
		buf := make([]byte, 1024)
		file, err := os.Open(infilename)
		if err != nil {
			panic(err)
		}
		io.CopyBuffer(out, file, buf)
		file.Close()
		waitCh <- true
	}()
}

// KillPid terminates a process and all processes belonging
// to the process group.
func KillPid(pid int) error {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return syscall.Kill(-pid, syscall.SIGKILL)
	}
	return syscall.Kill(-pgid, syscall.SIGKILL)
}

// SuspendPid stops a process group from its execution. Note
// that it sends SIGTSTP which can be caught by the application
// and hence could be ignored.
func SuspendPid(pid int) error {
	return syscall.Kill(-pid, syscall.SIGTSTP)
}

// ResumePid contiues to run a previously suspended process group.
func ResumePid(pid int) error {
	return syscall.Kill(-pid, syscall.SIGCONT)
}
