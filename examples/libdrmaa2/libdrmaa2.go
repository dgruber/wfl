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
		Args:          []string{"30"},
	}

	ctx := libdrmaa.NewLibDRMAAContextByCfgWithInitParams(libdrmaa.Config{
		DBFile:          "sessionmanager.db",
		DefaultTemplate: drmaa2interface.JobTemplate{},
	}, libdrmaa.LibDRMAASessionParams{
		UsePersistentJobStorage: true,
		DBFilePath:              "job.db",
	}).OnError(epanic)

	flow := wfl.NewWorkflow(ctx).OnError(epanic)

	// list running jobs (found in DB)
	for _, job := range flow.GetJobs() {
		fmt.Printf("found job %s in state %s\n", job.JobID(), job.State())
	}

	fmt.Println("submitting new job")
	job := flow.RunT(sleep)
}
