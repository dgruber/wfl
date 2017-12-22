package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

func Sh(command string) drmaa2interface.JobTemplate {
	return drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", command},
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stdout",
	}
}

func main() {
	job := wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContext()))
	job.RunT(Sh(`echo hello process`)).Wait()

	job = wfl.NewJob(wfl.NewWorkflow(wfl.NewDockerContextByCfg(wfl.DockerConfig{DefaultDockerImage: "golang:latest"})))
	job.RunT(Sh(`echo hello Docker`)).Wait()
}
