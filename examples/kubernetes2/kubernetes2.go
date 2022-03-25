package main

import (
	"fmt"

	"encoding/base64"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

func main() {

	flow := getKubernetesWorkflow()

	fmt.Println("Submitting a batch job to Kubernetes")

	jobTemplate := drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", `env && cat /input/data.txt`},
		JobEnvironment: map[string]string{
			"myenv":      "mycontent",
			"anotherenv": "anothercontent",
		},
	}
	// must be set separatly due to struct embedding
	jobTemplate.ExtensionList = map[string]string{
		// There must be a secret called my-credentials-secret.
		// To create one:
		// kubectl create secret generic my-credentials-secret --from-literal=password=secret
		"env-from-secret": "my-credentials-secret",
	}

	// Data can also be added as files into the container, the content
	// of the files can be stored as secrets or configmaps. The content
	// source is the base64 encoded string defined here or come from
	// existing sources.
	jobTemplate.StageInFiles = map[string]string{
		"/input/data.txt": "configmap-data:" +
			base64.StdEncoding.EncodeToString([]byte("\nmy input data set")),
	}

	job := flow.RunT(jobTemplate)
	if job.Errored() {
		fmt.Printf("Failed submitting job to Kubernetes")
	}

	fmt.Printf("Waiting for job %s to finish.\n", job.JobID())
	job.Wait()
	if job.State() == drmaa2interface.Failed {
		fmt.Printf("Job failed with exit code: %d\n", job.ExitStatus())
	}
	fmt.Printf("Job state: %s\n", job.State().String())

	// Print the output of the job
	jobInfo := job.JobInfo()
	if jobInfo.ExtensionList != nil {
		fmt.Printf("job output: %s\n", jobInfo.ExtensionList["output"])
	}

	fmt.Println("Removing job objects from Kubernetes")
	job.ReapAll()
}

func getKubernetesWorkflow() *wfl.Workflow {
	return wfl.NewWorkflow(kubernetes.NewKubernetesContextByCfg(
		kubernetes.Config{
			DefaultImage: "busybox:latest", // default container image Run() is using
			Namespace:    "default",        // must not be set as this is the default setting
		}))
}
