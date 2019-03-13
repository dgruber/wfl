package slurmcli

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"strings"
)

// conver into sbatch command line args
func convertJobTemplate(account string, jt drmaa2interface.JobTemplate) ([]string, error) {
	args := []string{"--parsable", "-A", account}
	if jt.RemoteCommand == "" {
		return nil, errors.New("RemoteCommand is not set")
	}
	args = append(args, jt.RemoteCommand)
	if jt.Args != nil && len(jt.Args) > 0 {
		args = append(args, jt.Args...)
	}
	return args, nil
}

func convertArrayJobArgs(start, end, step, maxParallel int) ([]string, error) {
	if maxParallel != 0 {
		return nil, fmt.Errorf("job array throttling is not supported in slurm")
	}
	arg := fmt.Sprintf("--array=%d-%d", start, end)
	if step != 1 {
		arg = fmt.Sprintf("%s:%d", arg, step)
	}
	return []string{arg}, nil
}

func convertState(state string) drmaa2interface.JobState {
	state = strings.TrimSpace(state)
	switch state {
	case "RUNNING":
		return drmaa2interface.Running
	case "COMPLETING":
		return drmaa2interface.Running
	case "COMPLETED":
		return drmaa2interface.Done
	case "CANCELLED":
		return drmaa2interface.Failed
	case "BOOT_FAIL":
		return drmaa2interface.Failed
	case "CONFIGURING":
		return drmaa2interface.Running
	case "DEADLINE":
		return drmaa2interface.Failed
	case "FAILED":
		return drmaa2interface.Failed
	case "NODE_FAIL":
		return drmaa2interface.Failed
	case "OUT_OF_MEMORY":
		return drmaa2interface.Failed
	case "PENDING":
		return drmaa2interface.Queued
	case "PREEMPTED":
		return drmaa2interface.Suspended
	case "RESV_DEL_HOLD":
		return drmaa2interface.QueuedHeld
	case "REQUEUE_FED":
		return drmaa2interface.Queued
	case "REQUEUE_HOLD":
		return drmaa2interface.RequeuedHeld
	case "REQUEUED":
		return drmaa2interface.Requeued
	case "RESIZING":
		return drmaa2interface.Running
	case "REVOKED":
		return drmaa2interface.Undetermined
	case "SIGNALING":
		return drmaa2interface.Failed
	case "SPECIAL_EXIT":
		return drmaa2interface.Requeued
	case "STAGE_OUT":
		return drmaa2interface.Running
	case "STOPPED":
		return drmaa2interface.Suspended
	case "SUSPENDED":
		return drmaa2interface.Suspended
	case "TIMEOUT":
		return drmaa2interface.Failed
	}
	fmt.Printf("slurm error: unknown state :%s\n", state)
	return drmaa2interface.Undetermined
}
