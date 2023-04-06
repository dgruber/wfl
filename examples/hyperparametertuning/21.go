package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
)

func main() {
	// Create a Docker context for running the training jobs.
	ctx := docker.NewDockerContextByCfg(
		docker.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				OutputPath:  wfl.RandomFileNameInTempDir() + "-{{ .ID }}",
				JobCategory: "parallel-training",
				ErrorPath:   "/dev/stderr",
			},
		},
	).WithSessionName("mnist-hyperparameter-tuning")

	flow := wfl.NewWorkflow(ctx)
	job := flow.NewJob()

	// Define the search space for hyperparameters.
	learningRates := []float64{0.001, 0.01, 0.1}
	batchSizes := []int{32, 64, 128} // 64, 128
	numEpochs := []int{10, 20, 30}   // 20, 30

	// Run training jobs concurrently with different combinations of hyperparameters.
	var bestAccuracy float64
	var bestParams string

	jobIDs := []string{}

	for _, lr := range learningRates {
		for _, bs := range batchSizes {
			for _, epochs := range numEpochs {
				// Run the training job with the current combination of hyperparameters.
				job.RunT(drmaa2interface.JobTemplate{
					RemoteCommand: "python3",
					Args: []string{"train_mnist.py",
						"--learning-rate",
						fmt.Sprintf("%f", lr), "--batch-size",
						strconv.Itoa(bs), "--num-epochs", strconv.Itoa(epochs)},
					JobEnvironment: map[string]string{
						"BATCH_SIZE":    strconv.Itoa(bs),
						"NUM_EPOCHS":    strconv.Itoa(epochs),
						"LEARNING_RATE": fmt.Sprintf("%f", lr),
					},
				}).OnError(func(err error) {
					panic(err)
				})
				jobIDs = append(jobIDs, job.JobID())

				// OK - we wait here for the job to finish. Let's not overwhelm the
				// macbook with too many jobs at the same time. We could do that when
				// submitting to an HPC cluster (libdrmaa) or a Cloud Provider (using
				// gcpbatchtracker)!
				job.Wait()
			}
		}
	}

	// Wait for all training jobs to complete.
	job.Synchronize()

	getJobOutput := func(j drmaa2interface.Job, i interface{}) error {
		if !IsJobIDInList(j.GetID(), jobIDs) {
			// Skip jobs that are not part of the hyperparameter tuning.
			return nil
		}
		// Get job output.
		template, err := j.GetJobTemplate()
		if err != nil {
			panic(err)
		}
		// Print Job OutputPath on stdout
		fmt.Printf("OutputPath: %s\n", template.OutputPath)
		// parse output from OutputPath
		content, err := ioutil.ReadFile(template.OutputPath)
		if err != nil {
			panic(err)
		}
		// Print output from the training job.
		fmt.Printf("Job %s: %s\n", j.GetID(), string(content))

		// get last line of content
		lines := strings.Split(string(content), "\n")
		lastLine := lines[len(lines)-2]
		// Parse the accuracy from the output.
		accuracy, err := strconv.ParseFloat(strings.TrimSpace(lastLine), 64)
		if err != nil {
			panic(err)
		}
		// Update the best accuracy and hyperparameters.
		if accuracy > bestAccuracy {
			bestAccuracy = accuracy
			bestParams = fmt.Sprintf("learning rate: %s, batch size: %s, num epochs: %s",
				template.JobEnvironment["LEARNING_RATE"],
				template.JobEnvironment["BATCH_SIZE"],
				template.JobEnvironment["NUM_EPOCHS"])
		}
		return nil
	}

	// Find the best hyperparameters.
	job.ForAll(getJobOutput, nil)

	// Print the best hyperparameters and accuracy.
	fmt.Printf("Best hyperparameters: %s\n", bestParams)
	fmt.Printf("Best accuracy: %.2f%%\n", bestAccuracy*100)
}

// IsJobIDInList returns true if the given job ID is in the list of job IDs.
func IsJobIDInList(jobID string, jobIDs []string) bool {
	for _, id := range jobIDs {
		if id == jobID {
			return true
		}
	}
	return false
}
