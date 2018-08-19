package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"os"
)

func main() {
	// build dgruber/job1:1.0.0 image
	buildImagesLocally()
	// run container
	runDockerWorkflow()
}

func buildImagesLocally() {
	pwf := wfl.NewWorkflow(wfl.NewProcessContext())
	pwf.RunT(newCompilerTemplate("job1", "dgruber")).
		OnError(defaultPanic()).
		OnFailure(defaultExit())
	pwf.RunT(newBuilderTemplate("job1", "dgruber")).
		OnError(defaultPanic()).
		OnFailure(defaultExit())
}

func newCompilerTemplate(job, owner string) drmaa2interface.JobTemplate {
	return drmaa2interface.JobTemplate{
		RemoteCommand: "./staging/compiler.sh",
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stdout",
		JobEnvironment: map[string]string{
			"job":        job,
			"owner":      owner,
			"image_name": job,
			"version":    "compiling",
		},
	}
}

func newBuilderTemplate(job, owner string) drmaa2interface.JobTemplate {
	return drmaa2interface.JobTemplate{
		RemoteCommand: "./staging/builder.sh",
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stdout",
		JobEnvironment: map[string]string{
			"job":        job,
			"owner":      owner,
			"image_name": job,
			"version":    "1.0.0",
		},
	}
}

func runDockerWorkflow() {
	wf := wfl.NewWorkflow(wfl.NewDockerContext())
	wf.RunT(drmaa2interface.JobTemplate{
		JobCategory:   "dgruber/job1:1.0.0",
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stdout",
		RemoteCommand: "/app/job",
	}).OnError(defaultPanic()).OnFailure(defaultExit())
}

func defaultPanic() func(error) {
	return func(e error) { panic(e) }
}

func defaultExit() func(drmaa2interface.Job) {
	return func(j drmaa2interface.Job) {
		fmt.Printf("job %s failed\n", j.GetID())
		os.Exit(1)
	}
}
