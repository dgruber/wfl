package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/libdrmaa"
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

	// Check https://github.com/dgruber/drmaa for compile time and
	// runtime requirements, like setting:
	//   export CGO_LDFLAGS="-L$SGE_ROOT/lib/$ARCH/"
	//   export CGO_CFLAGS="-I$SGE_ROOT/include"
	// and LD_LIBRARY_PATH for finding libdrmaa.so at runtime.

	sleep := drmaa2interface.JobTemplate{
		RemoteCommand:  "/bin/sh",
		Args:           []string{"-c", `echo sleeping $SECONDS second\(s\) && sleep $seconds && whoami`},
		OutputPath:     "/dev/stdout",
		ErrorPath:      "/dev/stderr",
		JobEnvironment: map[string]string{"SECONDS": "1"}, // environment variables set in container
	}

	ctx := libdrmaa.NewLibDRMAAContext().OnError(epanic)

	wf := wfl.NewWorkflow(ctx).OnError(epanic)

	job := wf.RunT(sleep).OnError(epanic).Do(start).OnSuccess(success).OnFailure(failure)

	job.Wait()
}
