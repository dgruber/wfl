package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/log"
)

func main() {

	flow := wfl.NewWorkflow(wfl.NewProcessContextByCfgWithInitParams(
		wfl.ProcessConfig{
			DBFile: "sessionmanager.db",
			DefaultTemplate: drmaa2interface.JobTemplate{
				OutputPath: "/dev/stdout",
				ErrorPath:  "/dev/stderr",
			},
		},
		simpletracker.SimpleTrackerInitParams{
			UsePersistentJobStorage: true,
			DBFilePath:              "job.db",
		},
	)).OnError(func(e error) { panic(e) })

	klogger, err := log.NewKlogLogger("WARNING")
	if err != nil {
		panic(err)
	}
	flow.SetLogger(klogger)

	for _, job := range flow.ListJobs() {
		fmt.Printf("found job %s in state %s\n", job.JobID(), job.State())
		if job.State() == drmaa2interface.Running {
			fmt.Printf("destroying running process\n")
			job.Kill()
		}
	}

	fmt.Printf("submitting a few jobs and exit; restart app to see see jobs and run more\n")
	flow.Run("sleep", "1").Resubmit(4).Synchronize()
}
