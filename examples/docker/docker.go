package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
)

func epanic(e error) { panic(e) }

func start(j drmaa2interface.Job) {
	fmt.Printf("Started job with ID: %s\n", j.GetID())
}

func success(j drmaa2interface.Job) {
	fmt.Printf("Job with ID %s finished successfully\n", j.GetID())
}

func failure(j drmaa2interface.Job) {
	ji, err := j.GetJobInfo()
	exit := -1
	if err == nil {
		exit = ji.ExitStatus
	}
	fmt.Printf("Job %s failed with exit status %d\n", j.GetID(), exit)
}

func main() {

	// JobName needs to be unique across calls as containers are not removed automatically!
	sleep := drmaa2interface.JobTemplate{
		//JobName:        "unique",
		RemoteCommand:  "/bin/sh",
		Args:           []string{"-c", `echo sleeping $seconds second\(s\) && sleep $seconds && whoami && ls /testdir`},
		JobCategory:    "golang:latest",                       // this is the docker image
		OutputPath:     "/dev/stdout",                         // stdout of container (here stdout of sconsole)
		ErrorPath:      "/dev/stderr",                         // stderr of container (here stderr of console)
		StageInFiles:   map[string]string{"/tmp": "/testdir"}, // mounts local tmp to /testdir in container
		JobEnvironment: map[string]string{"seconds": "1"},     // environment variables set in container
	}

	// Docker specific extensions to job template
	sleep.ExtensionList = map[string]string{
		"exposedPorts": "8124:8080/tcp", // ports redirected from container 8080 to local host 8124
		"user":         "root"}          // user name in container (needs to exist)

	// NewDockerContext() contacts with local docker when running.
	//
	// Optionally you can point to a specific Docker server by:
	//
	// "Use DOCKER_HOST to set the url to the docker server.
	//  Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
	//  Use DOCKER_CERT_PATH to load the TLS certificates from.
	//  Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default.", Moby, 2017
	//
	ctx := docker.NewDockerContext().OnError(epanic)

	wf := wfl.NewWorkflow(ctx).OnError(epanic)

	job := wf.RunT(sleep).OnError(epanic).Do(start).OnSuccess(success).OnFailure(failure)

	job.Wait()
}
