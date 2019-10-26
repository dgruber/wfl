package main

import (
	"errors"
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

// run on Linux with Singularity installed

var imageName = "alpine.simg"

var echo drmaa2interface.JobTemplate = drmaa2interface.JobTemplate{
	RemoteCommand: "echo",
	Args:          []string{"inside the container"},
	OutputPath:    "/dev/stdout",
	ErrorPath:     "/dev/stderr",
}

func main() {
	// build Singularity image
	if err := ImageBuilder(imageName, "SingularityRecipe"); err != nil {
		panic(err)
	}
	// run job workflow with the locally build image as default
	// image for the jobs if no other is specified in the JobCategory.
	RunWorkflow(imageName)
}

// ImageBuilder builds a new singularity container image from a recipe.
func ImageBuilder(image, recipe string) error {
	// Process context for starting plain OS processes.
	wf := wfl.NewWorkflow(wfl.NewProcessContext())

	// remove existing image to prevent issues when rebuilding
	wf.Run("rm", "alpine.simg").Wait()

	// build image
	job := wf.RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "sudo",
		Args:          []string{"singularity", "build", image, recipe},
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stderr",
	}).OnError(func(e error) {
		panic(e)
	}).Wait()

	if !job.Success() {
		return errors.New("building image failed")
	}
	return nil
}

// RunWorkflow runs a workflow of commands within different containers.
func RunWorkflow(image string) {
	// Singularity context to run commands within a Singularity container.
	wf := wfl.NewWorkflow(wfl.NewSingularityContextByCfg(wfl.SingularityConfig{DefaultImage: image}))

	wf.OnError(func(e error) {
		panic(e)
	})

	// run 100 Singularity containers in parallel and retry each failed one up to 3 times.
	job := wf.RunT(echo).
		Resubmit(99).
		Synchronize().
		RetryAnyFailed(3)

	if job.HasAnyFailed() {
		fmt.Printf("After 3 retries there are still failed jobs: %v\n", job.ListAllFailed())
		return
	}
}
