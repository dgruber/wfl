package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/libdrmaa"
)

func epanic(e error) { panic(e) }

func start(j drmaa2interface.Job) {
	fmt.Printf("New job started job with ID: %s\n", j.GetID())
}

func main() {

	// Check https://github.com/dgruber/drmaa for compile time and
	// runtime requirements, like setting:
	//   export CGO_LDFLAGS="-L$SGE_ROOT/lib/$ARCH/"
	//   export CGO_CFLAGS="-I$SGE_ROOT/include"
	// and LD_LIBRARY_PATH for finding libdrmaa.so at runtime.

	sleep := drmaa2interface.JobTemplate{
		JobName:       "testjob",
		RemoteCommand: "/bin/sleep",
		Args:          []string{"1"},
	}

	ctx := libdrmaa.NewLibDRMAAContext().OnError(epanic)

	wf := wfl.NewWorkflow(ctx).OnError(epanic)

	fmt.Println("submitting job and waiting for its end")
	job := wf.RunT(sleep).Wait()

	fmt.Printf("job state: %s\n", job.State())
	fmt.Printf("exit status: %d\n", job.ExitStatus())

	fmt.Printf("%v\n", job.JobInfo())

	if job.Success() {
		fmt.Println("succeeded")
	} else {
		fmt.Println("failed")
	}

	job.Wait()
}
