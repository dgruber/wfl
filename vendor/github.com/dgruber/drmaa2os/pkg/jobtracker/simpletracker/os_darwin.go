package simpletracker

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dgruber/drmaa2interface"
)

func OSStateStringforPID(pid string) (string, error) {
	/*
	   ps -v 85562

	   PID   STAT      TIME  SL  RE PAGEIN      VSZ    RSS   LIM     TSIZ  %CPU %MEM COMMAND
	   85562 S      0:00.00   0   0      0  2432788    640     -        0   0,0  0,0 /bin/sleep 123
	*/
	cmd := exec.Command("ps", "-v", pid)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected ps output (%s)", out)
	}
	elements := strings.Fields(lines[1])
	if len(elements) < 2 {
		return "", fmt.Errorf("unexpected amount of elements in ps output (%v)", elements)
	}
	return elements[1], nil
}

func OSStateToDRMAA2State(os string) drmaa2interface.JobState {

	if strings.Contains(os, "T") {
		return drmaa2interface.Suspended
	}

	// TODO not found

	return drmaa2interface.Running
}
