package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

// template specifies all default settings for jobs for the worklflow.
// That way Run() methods can be used in a container based (i.e. Singularity
// or Docker based) workflow where specifying the JobCategory is
// mandatory.
var template = drmaa2interface.JobTemplate{
	JobCategory: "shub://vsoch/hello-world:latest",
	OutputPath:  "/dev/stdout",
	ErrorPath:   "/dev/stdout",
}

var e = func(e error) { fmt.Printf("%s\n", e.Error()) }

func main() {
	flow := wfl.NewWorkflow(wfl.NewSingularityContextByCfg(wfl.SingularityConfig{
		DefaultTemplate: template,
	})).OnError(e)

	flow.Run("echo", "inside of container 1").OnError(e).ThenRun("echo", "inside of container 2").Wait()
}
