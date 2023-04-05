package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

func main() {
	SimpleProcessOutput()
	SimpleDockerOutput()
	// does not print to local /dev/stdout as output path
	SimpleKubernetesOutput()
}

func SimpleProcessOutput() {

	ctx := wfl.NewProcessContextByCfg(
		wfl.ProcessConfig{
			DefaultTemplate: drmaa2interface.JobTemplate{
				// {{ .ID }} is replaced by the task ID internally
				OutputPath: wfl.RandomFileNameInTempDir() + "-{{ .ID }}",
			},
		},
	).WithSessionName("wfl-example")

	runFlow(ctx)

}

func SimpleDockerOutput() {

	ctx := docker.NewDockerContextByCfg(
		docker.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				// {{ .ID }} is replaced by the task ID internally
				OutputPath:  wfl.RandomFileNameInTempDir() + "-{{ .ID }}",
				JobCategory: "alpine:latest",
			},
		},
	).WithSessionName("wfl-example")

	runFlow(ctx)
}

func SimpleKubernetesOutput() {

	ctx := kubernetes.NewKubernetesContextByCfg(
		kubernetes.Config{
			DefaultTemplate: drmaa2interface.JobTemplate{
				// {{ .ID }} is replaced by the task ID internally
				OutputPath:  wfl.RandomFileNameInTempDir() + "-{{ .ID }}",
				JobCategory: "alpine:latest",
			},
		},
	).WithSessionName("wfl-example")

	runFlow(ctx)
}

func runFlow(ctx *wfl.Context) {
	// Create a process that prints "Hello World" to stdout.
	// Internally it writes to an random output file, which
	// for each task in the flow adds a suffix with the task ID.
	// Output() waits until the job is finished and reads the
	// output file and returns the content.
	singleFileFlow := wfl.NewWorkflow(ctx).Run("echo", "Hello World")

	fmt.Printf("Output: %s\n", singleFileFlow.Output())

	fmt.Println(singleFileFlow.Run("echo", "some test").Output())

	// Print to console a second time, that does not work in Kubernetes
	// as the output path is forwarded to the calling process.
	ctx.DefaultTemplate.OutputPath = "/dev/stdout"

	wfl.NewWorkflow(ctx).Run("echo", singleFileFlow.Output()).
		Wait().
		ReapAll()
}
