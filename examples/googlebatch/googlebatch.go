package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/googlebatch"
)

func main() {

	// set privileges: gcloud auth application-default login
	flow := wfl.NewWorkflow(googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultJobCategory: googlebatch.JobCategoryScript, // default container image Run() is using or script if cmd runs as script
			GoogleProjectID:    "customer-nest",
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
	fmt.Printf("running job in Google Batch\n")
	job := flow.Run(`echo hello google batch`)
	if job.Errored() {
		panic(job.LastError())
	}
	fmt.Printf("waiting for job to be finished")
	job.Wait()
}
