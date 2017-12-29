package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/jstream"
)

func step1(j *wfl.Job) *wfl.Job {
	fmt.Printf("Started job %s\n", j.JobID())
	return j
}

func step2(j *wfl.Job) *wfl.Job {
	fmt.Printf("Processing job %s\n", j.JobID())
	return j
}

func exitStatusSmallerEqualsTen(j *wfl.Job) bool {
	return j.ExitStatus() <= 10
}

func main() {

	template := wfl.NewTemplate(drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", `exit $TASK_ID`},
	}).AddIterator("tasks", wfl.NewEnvSequenceIterator("TASK_ID", 1, 1))

	cfg := jstream.Config{
		Template:   template,
		Workflow:   wfl.NewWorkflow(wfl.NewProcessContext()),
		BufferSize: 16,
	}

	jobs := jstream.NewStream(cfg, jstream.NewSequenceBreaker(100)).
		Apply(step1).
		Apply(step2).
		Synchronize().
		Filter(exitStatusSmallerEqualsTen).
		Collect()

	fmt.Printf("Amount of jobs with exit status <= 10: %d\n", len(jobs))
}
