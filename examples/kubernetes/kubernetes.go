package main

import (
	"fmt"

	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

func main() {
	wf := wfl.NewWorkflow(kubernetes.NewKubernetesContextByCfg(
		kubernetes.Config{
			DefaultImage: "golang:latest",
		}))
	fmt.Println("Submitting 10 sleep batch jobs to kubernetes")
	if wf.Run("sleep", "5").Resubmit(9).Synchronize().AnyFailed() {
		fmt.Println("Not all jobs run successfully")
	} else {
		fmt.Println("All 10 sleep jobs finished successfully")
	}
}
