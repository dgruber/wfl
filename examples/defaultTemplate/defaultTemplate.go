package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

// template specifies all default settings for jobs for the worklflow.
// That way Run() methods can be used in a container based (i.e. Singularity
// or Docker based) workflow where specifying the JobCategory is
// mandatory.
var template = drmaa2interface.JobTemplate{
	JobCategory: "alpine:latest",
	OutputPath:  "/dev/stdout",
	ErrorPath:   "/dev/stderr",
}

func main() {
	flow := wfl.NewWorkflow(wfl.NewDockerContextByCfg(wfl.DockerConfig{
		DefaultTemplate: template,
	}))

	flow.Run("echo", "inside of container 1").ThenRun("echo", "inside of container 2").Wait()

}
