package main

import (
	"fmt"

	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

func main() {

	// open a connection to the Kubernetes cluster
	flow := wfl.NewWorkflow(kubernetes.NewKubernetesContextByCfg(
		kubernetes.Config{
			DefaultImage: "busybox:latest", // default container image Run() is using
			Namespace:    "default",        // must not be set as this is the default setting
		}))

	fmt.Println("Submitting 5 sleep batch jobs to Kubernetes")
	job := flow.Run("/bin/sh", "-c", `exit $(($RANDOM%2))`).Resubmit(4)

	// for more flexibility you can use RunT() with all what the DRMAA2
	// job template for Kubernetes offers (see https://github.com/dgruber/drmaa2os)

	fmt.Println("Waiting for all jobs to be finished and check for job failure")
	if job.Synchronize().AnyFailed() {
		for _, failed := range job.ListAllFailed() {
			// failed is a DRMAA2 job object
			jinfo, err := failed.GetJobInfo()
			if err != nil {
				fmt.Printf("failed to get JobInfo of job %s: %v\n", failed.GetID(), err)
				continue
			}
			fmt.Printf("Job %s failed with exit status %d.\n",
				failed.GetID(), jinfo.ExitStatus)
		}
	} else {
		fmt.Println("All 5 sleep jobs finished successfully")
	}

	fmt.Println("Removing job objects from Kubernetes")
	job.ReapAll()
}
