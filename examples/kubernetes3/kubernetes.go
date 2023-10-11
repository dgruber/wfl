package main

import (
	"fmt"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/extension"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

func main() {
	flow := getKubernetesWorkflow()

	// default job handler:
	// - prints success or failure message when the job finished
	// - panics on job submission error
	// Override success handler to print start and runtime of the job
	observer := wfl.NewDefaultObserver()
	observer.SuccessHandler = func(j drmaa2interface.Job) {
		fmt.Printf("job %s finished successfully\n", j.GetID())
		ji, err := j.GetJobInfo()
		if err != nil {
			fmt.Printf("failed to get job info: %v\n", err)
		} else {
			fmt.Printf("job %s start time: %s\n", j.GetID(), ji.DispatchTime)
			fmt.Printf("job %s run time: %s\n", j.GetID(), ji.WallclockTime.Round(time.Millisecond))
			fmt.Printf("job %s run on following machines: %v\n", j.GetID(), ji.AllocatedMachines)
		}
	}

	job := flow.RunT(echo("hello world")).Observe(observer)

	output := getJobOutput(job)

	job.RunT(toUpper(output)).Observe(observer)

	output = getJobOutput(job)

	fmt.Printf("output: %s\n", output)

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

func getJobOutput(job *wfl.Job) string {
	jinfo := job.JobInfo()
	output, _ := jinfo.ExtensionList[extension.JobInfoK8sJSessionJobOutput]
	return output
}

func echo(input string) drmaa2interface.JobTemplate {
	return drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", `echo $MY_INPUT_DATA $MY_INPUT_DATA`},
		JobEnvironment: map[string]string{
			"MY_INPUT_DATA": input,
		},
	}
}

func toUpper(input string) drmaa2interface.JobTemplate {
	return drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", `echo $MY_INPUT_DATA | tr 'a-z' 'A-Z'`},
		JobEnvironment: map[string]string{
			"MY_INPUT_DATA": input,
		},
	}
}
