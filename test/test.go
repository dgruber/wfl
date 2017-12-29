package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"os"
)

// Compiles all test applications
// ------------------------------
//
// - with local go compiler
// - inside a Docker container (golang:latest needs to be pulled before)

// works on darwin / linux

var exitCode int

func createProcBuild() (map[string]string, drmaa2interface.JobTemplate, *wfl.Job) {

	testApps := map[string]string{
		"simple":       "../examples/simple/simple.go",
		"touchy":       "../examples/touchy/touchy.go",
		"cloudfoundry": "../examples/cloudfoundry/cloudfoundry.go",
		"docker":       "../examples/docker/docker.go",
		"template":     "../examples/template/template.go",
		"parallel":     "../examples/parallel/parallel.go",
		"notifier":     "../examples/notifier/notifier.go",
		"shell":        "../examples/shell/shell.go",
		"stream":       "../examples/stream/stream.go",
	}

	jtemplate := drmaa2interface.JobTemplate{
		RemoteCommand: "go",
		Args:          []string{"build", "-a"},
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stderr",
	}

	job := wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContext()))

	return testApps, jtemplate, job
}

func createDockerBuild(image string) (map[string]string, drmaa2interface.JobTemplate, *wfl.Job) {

	testApps := map[string]string{
		"simple":       "/go/src/github.com/dgruber/wfl/examples/simple/simple.go",
		"touchy":       "/go/src/github.com/dgruber/wfl/examples/touchy/touchy.go",
		"cloudfoundry": "/go/src/github.com/dgruber/wfl/examples/cloudfoundry/cloudfoundry.go",
		"docker":       "/go/src/github.com/dgruber/wfl/examples/docker/docker.go",
		"template":     "/go/src/github.com/dgruber/wfl/examples/template/template.go",
		"parallel":     "/go/src/github.com/dgruber/wfl/examples/parallel/parallel.go",
		"shell":        "/go/src/github.com/dgruber/wfl/examples/shell/shell.go",
		"stream":       "/go/src/github.com/dgruber/wfl/examples/stream/stream.go",
	}

	goPath := os.Getenv("GOPATH")

	jtemplate := drmaa2interface.JobTemplate{
		RemoteCommand: "go",
		Args:          []string{"build", "-a"},
		JobCategory:   image,
		StageInFiles: map[string]string{
			goPath + "/src/github.com/dgruber/drmaa2interface": "/go/src/github.com/dgruber/drmaa2interface",
			goPath + "/src/github.com/dgruber/wfl":             "/go/src/github.com/dgruber/wfl"},
	}

	ctx := wfl.NewDockerContextByCfg(wfl.DockerConfig{DefaultDockerImage: image})
	if ctx.HasError() {
		fmt.Printf("Docker context not supported: %s\n", ctx.Error())
		testApps = nil
	}

	wf := wfl.NewWorkflow(ctx)
	if wf.HasError() {
		fmt.Printf("Docker workflow not supported: %s\n", ctx.Error())
		testApps = nil
	}

	job := wfl.NewJob(wf)

	return testApps, jtemplate, job
}

func executeWorkflow(testApps map[string]string, jtemplate drmaa2interface.JobTemplate, job *wfl.Job) {
	orignalArgs := jtemplate.Args
	for app, path := range testApps {
		jtemplate.Args = append(orignalArgs, path)
		job.RunT(jtemplate).Do(func(j drmaa2interface.Job) {
			fmt.Printf("Building %s (%s)\n", app, j.GetID())
		}).OnSuccess(func(j drmaa2interface.Job) {
			fmt.Printf("%s build successfully\n", app)
		}).OnFailure(func(j drmaa2interface.Job) {
			fmt.Printf("failed building %s\n", app)
			exitCode = 1
		}).OnError(func(err error) {
			fmt.Printf("error: %s\n", err)
		})
	}
}

func main() {
	fmt.Println("Building examples in local processes using local go compiler")
	executeWorkflow(createProcBuild())

	fmt.Println("Building examples in golang:latest Docker containers")
	executeWorkflow(createDockerBuild("golang:latest"))

	os.Exit(exitCode)
}
