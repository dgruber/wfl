package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

func main() {

	sleep := drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"1"},
		JobCategory:   "golang", // this is the docker image
	}

	// Optionally you can point to a specific Docker server by:
	//
	// "Use DOCKER_HOST to set the url to the docker server.
	//  Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
	//  Use DOCKER_CERT_PATH to load the TLS certificates from.
	//  Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default.", Moby, 2017
	//
	ctx := wfl.NewDockerContext().OnError(func(e error) { panic(e) })

	wfl.NewWorkflow(ctx).OnError(func(e error) {
		panic("error during workflow creation " + e.Error())
	}).RunT(sleep).Do(func(j drmaa2interface.Job) {
		fmt.Printf("Started job with ID: %s\n", j.GetID())
	}).OnSuccess(func(j drmaa2interface.Job) {
		fmt.Println("Job finished successfully")
	})

	// when setting golang as default Docker image JobCategory is not required to
	// be set and the simplified Run() methods can be used.

	ctx2 := wfl.NewDockerContextByCfg(wfl.DockerConfig{DefaultDockerImage: "golang"})

	wfl.NewWorkflow(ctx2).OnError(func(e error) {
		panic("error during workflow creation " + e.Error())
	}).Run("sleep", "1").Do(func(j drmaa2interface.Job) {
		fmt.Printf("Started job with ID: %s\n", j.GetID())
	}).OnSuccess(func(j drmaa2interface.Job) {
		fmt.Println("Job finished successfully")
	})
}
