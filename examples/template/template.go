package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

// note that is written and executed on macOS

// Demonstration of how Template can be used for generating a sequence
// of environment variables (TASK_ID) in the job context so that each
// job can process a different data chunk by translating the content of
// the TASK_ID environment variable to the data to process.

// implements a canonical example of a workflow where the first task
// does some pre-processing (here generating a sequence of characters),
// then each data chunk is processed in parallel (here each character
// is converted from lowercase to uppercase), and finally when all
// jobs finished the output files are combined.

func main() {

	// create a Template with an interator which adds TASK_ID as
	// environment variable setting it to 1 and increments it
	// for each job (each time .Next() is called).
	template := wfl.NewTemplate(drmaa2interface.JobTemplate{
		RemoteCommand: "./scripts/convert.sh",
		Args:          []string{"./input/input.txt"},
	}).AddIterator("files", wfl.NewEnvSequenceIterator("TASK_ID", 1, 1))

	// create an empty job to begin to work with OS processes
	job := wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContext()))

	// create a random lowercase string with 5 characters, store it
	// in input.txt and wait for job to finish
	job.Run("./scripts/create_input.sh", "5", "./input/input.txt").Wait()

	// run jobs which converts a character in uppercase in parallel
	for i := 0; i < 5; i++ {
		job.RunT(template.Next())
	}

	// wait for all jobs and combine the output files to output.txt
	job.Synchronize().Run("./scripts/combine_output.sh", "5", "./output/output.txt").Wait()

}
