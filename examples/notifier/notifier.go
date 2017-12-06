package main

import (
	"fmt"
	"github.com/dgruber/wfl"
	"os"
)

func main() {

	notifier := wfl.NewNotifier()

	go func() {
		// pre-proc followed by parallel exec
		cfg := wfl.ProcessConfig{
			DBFile: "wf1.db",
		}
		defer os.Remove("wf1.db")

		wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContextByCfg(cfg))).
			TagWith("A").
			Run("sleep", "1").
			ThenRun("sleep", "3").
			Run("sleep", "2").
			Synchronize().
			Notify(notifier)
	}()

	go func() {
		// pre-proc followed by parallel exec
		cfg := wfl.ProcessConfig{
			DBFile: "wf2.db",
		}
		defer os.Remove("wf2.db")

		wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContextByCfg(cfg))).
			TagWith("B").
			Run("sleep", "1").
			ThenRun("sleep", "2").
			Run("sleep", "2").
			Synchronize().
			Notify(notifier)
	}()

	job1 := notifier.ReceiveJob()
	fmt.Printf("finished with sequence: %s\n", job1.Tag())

	job2 := notifier.ReceiveJob()
	fmt.Printf("finished with sequence: %s\n", job2.Tag())

}
