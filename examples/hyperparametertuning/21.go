package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
	"golang.org/x/exp/slices"
)

func main() {
	// Create a Docker context for running the training jobs.
	ctx := docker.NewDockerContextByCfg(
		docker.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				OutputPath:  wfl.RandomFileNameInTempDir(),
				JobCategory: "parallel-training",
				ErrorPath:   "/dev/stderr",
			},
		},
	).WithUniqueSessionName()

	flow := wfl.NewWorkflow(ctx)
	job := flow.NewJob()

	// Define the search space for hyperparameters.
	learningRates := []float64{0.001, 0.01, 0.1}
	batchSizes := []int{32, 64, 128}
	numEpochs := []int{10, 20, 30}

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
				// Remove that line if you want to run all jobs in parallel.
				job.Wait()
			}
		}
	}

	// Wait for all training jobs to complete.
	job.Synchronize()

	getJobOutput := func(j drmaa2interface.Job, i interface{}) error {
		if !slices.Contains(jobIDs, j.GetID()) {
			// Skip jobs that are not part of the hyperparameter tuning.
			return nil
		}
		// Get job output.
		template, err := j.GetJobTemplate()
		if err != nil {
			panic(err)
		}
		fmt.Printf("OutputPath: %s\n", template.OutputPath)
		content, err := ioutil.ReadFile(template.OutputPath)
		if err != nil {
			panic(err)
		}
		// Print output from the training job.
		fmt.Printf("Job %s: %s\n", j.GetID(), string(content))

		// Get last line of content
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
	job.ForEach(getJobOutput, nil)

	// Print the best hyperparameters and accuracy.
	fmt.Printf("Best hyperparameters: %s\n", bestParams)
	fmt.Printf("Best accuracy: %.2f%%\n", bestAccuracy*100)
}
