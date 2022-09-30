package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/gcpbatchtracker"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/googlebatch"
)

func main() {

	// set privileges on cli before running: "gcloud auth application-default login"

	// for job template -> google batch mappings please check:
	// https://github.com/dgruber/gcpbatchtracker

	flow := wfl.NewWorkflow(googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultJobCategory: googlebatch.JobCategoryScript, // default container image Run() is using or script if cmd runs as script
			GoogleProjectID:    "YOURGOOGLEPROJECTID",         // be sure to set the correct project id
			Region:             "europe-north1",
			DefaultTemplate: drmaa2interface.JobTemplate{
				MinSlots: 1, // for MPI set MinSlots = MaxSlots and > 1
				MaxSlots: 1, // for just a bund of tasks MinSlots = 1 (parallelism) and MaxSlots = <tasks>
			},
		}))
	if flow.HasError() {
		// there was an error creating the workflow context
		panic(flow.Error())
	}

	fmt.Println("Submitting job to Google Batch...")

	// using default settings inherited from flow
	job := flow.Run(`echo hello google batch`)
	if job.Errored() {
		panic(job.LastError())
	}

	fmt.Println("Submitting job with more fine grained control...")

	// override defaults with RunT()
	job = job.RunT(drmaa2interface.JobTemplate{
		CandidateMachines: []string{"n2-highcpu-2"}, // machine type
		MinSlots:          1,                        // this is parallelism
		MaxSlots:          2,                        // this is task count
		RemoteCommand:     "/bin/bash",
		Args:              []string{"-c", `hello from google batch task ${BATCH_TASK_INDEX}`},
		Extension: drmaa2interface.Extension{
			ExtensionList: map[string]string{
				gcpbatchtracker.ExtensionSpot: "true",
				gcpbatchtracker.ExtensionProlog: `#!/bin/bash
echo "hello from prolog"
`,
			},
		},
	})
	if job.Errored() {
		panic(job.LastError())
	}

	fmt.Println("Waiting for all jobs to be finished...")
	job.Synchronize()

	fmt.Printf("Job finished with state %s\n", job.State().String())

}
