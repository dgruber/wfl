package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/gcpbatchtracker"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/googlebatch"
)

var (
	GoogleProject string
	GCPBucketName string
)

func main() {

	GoogleProject = os.Getenv("GOOGLE_PROJECT")
	if GoogleProject == "" {
		panic("GOOGLE_PROJECT environment variable not set")
	}

	GCPBucketName = os.Getenv("GOOGLE_BUCKET")
	if GCPBucketName == "" {
		panic("GOOGLE_BUCKET environment variable not set")
	}

	fmt.Printf("Using Google project: %s\n", GoogleProject)
	fmt.Printf("Using Google bucket: %s\n", GCPBucketName)

	// Create a Docker context for running the training jobs.
	ctx := googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				JobCategory:       "gcr.io/" + GoogleProject + "/cifar10-parallel-training",
				CandidateMachines: []string{"n1-standard-16"},
				MinSlots:          1,
				MaxSlots:          1,
			},
			GoogleProjectID: GoogleProject,
			Region:          "us-central1",
		},
	).WithUniqueSessionName()

	numJobs := 4 // Number of parallel training jobs

	flow := wfl.NewWorkflow(ctx)

	// Run the data preparation job.
	dataPrepJob := flow.RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "python",
		Args: []string{
			"prepare_cifar10.py",
			"--num-parts", fmt.Sprintf("%d", numJobs),
			"--output-dir", "/output",
		},
		JobEnvironment: map[string]string{
			"NUM_PARTS":     fmt.Sprintf("%d", numJobs), // Number of dataset parts to create
			"GOOGLE_BUCKET": GCPBucketName,
		},
		StageInFiles: map[string]string{
			"/output": "gs://" + GCPBucketName,
		},
		Extension: drmaa2interface.Extension{
			ExtensionList: map[string]string{
				"spot": "true",
			},
		},
	}).OnError(func(err error) {
		panic(err)
	})

	fmt.Printf("Data preparation job submitted: %s\n", dataPrepJob.JobID())

	// Wait for the data preparation job to complete.
	dataPrepJob.Wait().OnError(func(err error) {
		panic(err)
	}).OnFailure(func(job drmaa2interface.Job) {
		fmt.Printf("Data preparation job failed: %s\n", job.GetID())
		os.Exit(1)
	}).OnSuccess(func(job drmaa2interface.Job) {
		fmt.Println("Data preparation job completed successfully")
		// print runtime of job
		ji, err := job.GetJobInfo()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Job runtime: %s\n", ji.WallclockTime.Round(time.Second).String())
	})

	job := flow.NewJob()

	// Run training jobs concurrently with different parts of the dataset.
	trainingJobIDs := make([]string, 0, numJobs)
	for i := 1; i <= numJobs; i++ {
		inputFile := fmt.Sprintf("cifar10_part%d.npz", i)
		outputFile := fmt.Sprintf("cifar10_trained_model_part%d.h5", i)
		accuracyFile := fmt.Sprintf("cifar10_accuracy_part%d.txt", i)
		logFile := fmt.Sprintf("cifar10_training_log_part%d.txt", i)

		// Run the training job with the current part of the dataset.
		job.RunT(drmaa2interface.JobTemplate{
			RemoteCommand: "python3",
			Args: []string{
				"train_cifar10.py",
				"--input-file", "/input/" + inputFile,
				"--output-file", "/input/" + outputFile,
				"--accuracy-file", "/input/" + accuracyFile,
				"--validation-data", "/input/cifar10_validation.npz",
			},
			JobEnvironment: map[string]string{
				"ACCURACY": accuracyFile,
				"PART":     fmt.Sprintf("%d", i),
				"LOGFILE":  logFile,
			},
			OutputPath: "gs://" + GCPBucketName + "/" + logFile,
			// Copy the input and output files to/from the Bucket.
			StageInFiles: map[string]string{
				"/input": "gs://" + GCPBucketName,
			},
			Extension: drmaa2interface.Extension{
				ExtensionList: map[string]string{
					"spot": "true",
					// add this using GPU support
					//"accelerators": "1*nvidia-tesla-v100",
				},
			},
		}).OnError(func(err error) {
			panic(err)
		})
		fmt.Printf("Submitted job %s: using dataset part %d\n", job.JobID(), i)
		trainingJobIDs = append(trainingJobIDs, job.JobID())
	}

	// Wait for all training jobs to complete.
	fmt.Printf("Waiting for %d jobs to complete...\n", numJobs)
	job.Synchronize()

	// Print the accuracy of each training job.
	job.ForEach(getJobOutput, trainingJobIDs)

}

var getJobOutput = func(j drmaa2interface.Job, i interface{}) error {
	trainingJobIDs := i.([]string)
	jobID := j.GetID()
	if !IsJobIDInList(jobID, trainingJobIDs) {
		return nil
	}
	jt, err := j.GetJobTemplate()
	if err != nil {
		return err
	}
	if j.GetState() == drmaa2interface.Failed {
		fmt.Printf("Job %s failed\n", j.GetID())
		// print log of the job
		out, err := gcpbatchtracker.ReadFromBucket("gs://"+GCPBucketName, jt.JobEnvironment["LOGFILE"])
		if err != nil {
			return err
		}
		fmt.Printf("Log of failed job %s:", j.GetID())
		fmt.Println(out)
		return errors.New("job failed")
	}
	fmt.Printf("Job %s completed successfully\n", j.GetID())

	out, err := gcpbatchtracker.ReadFromBucket("gs://"+GCPBucketName, jt.JobEnvironment["ACCURACY"])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Accuracy of model trained with part %s: %s\n",
		jt.JobEnvironment["ACCURACY"], out)
	info, err := j.GetJobInfo()
	if err != nil {
		return err
	}
	fmt.Printf("Job runtime: %s\n", info.WallclockTime.Round(time.Second).String())
	return nil
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
