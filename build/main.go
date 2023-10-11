package main

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
)

func main() {

	flow := wfl.NewWorkflow(docker.NewDockerContextByCfg(
		docker.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				JobCategory: "golang:aline",
				OutputPath:  "/dev/stdout",
				ErrorPath:   "/dev/stderr",
			},
		}))

	flow.Run("go build -a").Wait()

}
