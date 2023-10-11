package main

import (
	"fmt"
	"os"

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

	GCPBucketName := os.Getenv("GOOGLE_BUCKET")
	if GCPBucketName == "" {
		panic("GOOGLE_BUCKET environment variable not set")
	}

	// Upload content and style image to Google Cloud Storage.
	gcpbatchtracker.CopyFileToBucket("gs://"+GCPBucketName,
		"content_image.jpg", "content_image.jpg")
	gcpbatchtracker.CopyFileToBucket("gs://"+GCPBucketName,
		"style_image.jpg", "style_image.jpg")

	// Create a Docker context for running the style transfer job.
	ctx := googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				JobCategory:       "gcr.io/" + googleProject + "/style-transfer",
				CandidateMachines: []string{"c2d-standard-8"},
				MinSlots:          1,
				MaxSlots:          1,
			},
			GoogleProjectID: googleProject,
			Region:          "us-central1",
		},
	).WithSessionName("style-transfer")

	// Run the style transfer job.
	job := wfl.NewWorkflow(ctx).NewJob()

	job.RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "python",
		Args: []string{
			"/style_transfer.py",
			"--content-image", "/bucket/content_image.jpg",
			"--style-image", "/bucket/style_image.jpg",
			"--output-image", "/bucket/styled_image.jpg",
		},
		StageInFiles: map[string]string{
			"/bucket": "gs://" + GCPBucketName,
		},
	}).OnError(func(err error) {
		panic(err)
	})

	fmt.Printf("Style transfer job submitted: %s\n", job.JobID())

	// Wait for the style transfer job to complete.
	job.Wait()

	if job.Success() {
		fmt.Println("Style transfer job completed successfully.")
	} else {
		fmt.Println("Style transfer job failed.")
		os.Exit(1)
	}

	outputImage, err := gcpbatchtracker.ReadFromBucket("gs://"+GCPBucketName,
		"styled_image.jpg")
	if err != nil {
		panic(err)
	}
	// write output image to disk
	f, err := os.Create("styled_image.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(outputImage)

	fmt.Println("Style transfer job completed. Output image (styled_image.jpg) written to disk.")
}
