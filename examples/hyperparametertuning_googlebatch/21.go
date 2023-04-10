package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/gcpbatchtracker"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/googlebatch"
)

func main() {

	googleProject := os.Getenv("GOOGLE_PROJECT")
	if googleProject == "" {
		panic("GOOGLE_PROJECT environment variable not set")
	}

	// Create a Docker context for running the training jobs.
	ctx := googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				JobCategory:       "gcr.io/" + googleProject + "/parallel-training",
				CandidateMachines: []string{"c2d-standard-8"},
				MinSlots:          1,
				MaxSlots:          1,
			},
			GoogleProjectID: googleProject,
			Region:          "us-central1",
		},
	).WithSessionName("mnist-hyperparameter-tuning")

	// Please adapt the bucket name to your needs.
	GCPBucketName := os.Getenv("GOOGLE_BUCKET")
	if GCPBucketName == "" {
		panic("GOOGLE_BUCKET environment variable not set")
	}

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

	id := 0
	for _, lr := range learningRates {
		for _, bs := range batchSizes {
			for _, epochs := range numEpochs {
				id++
				outputFile := fmt.Sprintf("mnist_hyperparameter_tuning_output_%d", id)
				outputPath := "/output/" + outputFile
				// Run the training job with the current combination of hyperparameters.
				job.RunT(drmaa2interface.JobTemplate{
					RemoteCommand: "python3",
					Args: []string{"train_mnist.py",
						"--learning-rate",
						fmt.Sprintf("%f", lr), "--batch-size",
						strconv.Itoa(bs), "--num-epochs", strconv.Itoa(epochs),
						"--output-file", outputPath},
					JobEnvironment: map[string]string{
						"BATCH_SIZE":    strconv.Itoa(bs),
						"NUM_EPOCHS":    strconv.Itoa(epochs),
						"LEARNING_RATE": fmt.Sprintf("%f", lr),
						"OUTPUT_FILE":   outputFile,
					},
					// Copy the output file to the Bucket.
					StageOutFiles: map[string]string{
						"/output": GCPBucketName,
					},
				}).OnError(func(err error) {
					panic(err)
				})
				fmt.Printf("Submitted job %s: learning rate: %f, batch size: %d, num epochs: %d\n",
					job.JobID(), lr, bs, epochs)
				jobIDs = append(jobIDs, job.JobID())
			}
		}
	}

	// Wait for all training jobs to complete.
	fmt.Printf("Waiting for %d jobs to complete...\n", len(jobIDs))
	job.Synchronize()

	getJobOutput := func(j drmaa2interface.Job, i interface{}) error {
		if !IsJobIDInList(j.GetID(), jobIDs) {
			// Skip jobs that are not part of the hyperparameter tuning.
			return nil
		}

		template, err := j.GetJobTemplate()
		if err != nil {
			return fmt.Errorf("could not get job template for job %s: %w",
				j.GetID(), err)
		}

		outputFileName := template.JobEnvironment["OUTPUT_FILE"]

		// Get job output from Bucket.
		output, err := gcpbatchtracker.ReadFromBucket("gs://"+GCPBucketName, outputFileName)
		if err != nil {
			return fmt.Errorf("could not get job output for job %s: %w",
				j.GetID(), err)
		}

		// Print output from the training job.
		fmt.Printf("Job %s: %s\n", j.GetID(), string(output))

		// Parse the accuracy from the output.
		accuracy, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
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
	err := job.ForAll(getJobOutput, nil)
	if err != nil {
		fmt.Printf("Error during job output retrieval: %s\n", err)
	}

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
