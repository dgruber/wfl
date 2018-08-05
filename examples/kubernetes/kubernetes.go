package main

import (
	"fmt"
	"github.com/dgruber/wfl"
)

func main() {
	wf := wfl.NewWorkflow(wfl.NewKubernetesContextByCfg(
		wfl.KubernetesConfig{
			DefaultImage: "golang:latest",
		}))
	fmt.Println("Submitting 10 sleep batch jobs to kubernetes")
	if wf.Run("sleep", "5").Resubmit(9).Synchronize().AnyFailed() {
		fmt.Println("Not all jobs run successfully")
	} else {
		fmt.Println("All 10 sleep jobs finished successfully")
	}
}
