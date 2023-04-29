package main

import (
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/kubernetes"
)

//go:embed bark.py
var bark string

func main() {

	googleProject := os.Getenv("GOOGLE_PROJECT")
	if googleProject == "" {
		panic("GOOGLE_PROJECT environment variable not set")
	}

	filestoreIPAddress := os.Getenv("FILESTORE_IP_ADDRESS")
	if filestoreIPAddress == "" {
		panic("FILESTORE_IP_ADDRESS environment variable not set")
	}

	filestoreName := os.Getenv("FILESTORE_NAME")
	if filestoreName == "" {
		panic("FILESTORE_NAME environment variable not set")
	}

	fmt.Printf("Using Google project: %s\n", googleProject)
	fmt.Printf("Using Google filestore IP address: %s\n", filestoreIPAddress)
	fmt.Printf("Using Google filestore name: %s\n", filestoreName)

	outputFileName := "my_output.wav"

	script := CreateBarkInferenceScript(123, outputFileName, ExamplePrompts())

	flow := wfl.NewWorkflow(kubernetes.NewKubernetesContextByCfg(
		kubernetes.Config{
			DefaultImage: "gcr.io/" + googleProject + "/bark",
			Namespace:    "default",
		})).OnErrorPanic()

	jobTemplate := CreateJobTemplate(script, filestoreIPAddress, filestoreName)

	job := flow.RunT(jobTemplate).OnErrorPanic()

	fmt.Printf("Waiting for job %s to finish. This can take a couple of minutes.\n",
		job.JobID())

	job.Wait()

	fmt.Printf("Job took %f seconds\n", job.JobInfo().WallclockTime.Seconds())
	fmt.Printf("Job output:\n%s\n", job.Output())

	if job.State() == drmaa2interface.Done {
		fmt.Printf("Sound file copied to /bark/%s\n", outputFileName)
	} else {
		fmt.Printf("Job failed with exit code: %d\n", job.ExitStatus())
	}

	// If you want to remove the finished job from Kubernetes uncomment the
	// following lines:
	//
	// fmt.Println("Removing job objects from Kubernetes.")
	// job.ReapAll()
}

func CreateJobTemplate(pythonScript, filestoreIPAddress, filestoreName string) drmaa2interface.JobTemplate {
	jobTemplate := drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/bash",
		Args: []string{
			"-c",
			"nvidia-smi && cd /home/bark/bark && python3 -c '" + pythonScript + "'",
		},
		WorkingDirectory: "/home/bark",
		MinSlots:         1,
		MaxSlots:         1,
		OutputPath:       "/tmp/joboutput",
		//ErrorPath:         "/tmp/joboutput",
		JobEnvironment: map[string]string{
			"NVIDIA_VISIBLE_DEVICES":     "all",
			"NVIDIA_DRIVER_CAPABILITIES": "compute,utility",
			"NVIDIA_REQUIRE_CUDA":        "cuda>=11.0",
		},
		StageInFiles: map[string]string{
			// see also: https://github.com/dgruber/drmaa2os/blob/master/pkg/extension/jobtemplate.go
			"/home/bark/output": "nfs:" + filestoreIPAddress + ":/" + filestoreName + "/bark",
		},
	}
	jobTemplate.ExtensionList = map[string]string{
		// this sets nodeSelector: cloud.google.com/gke-accelerator: nvidia-tesla-v100
		// and resource limits to nvidia.com/gpu: 1
		"distribution": "gke",
		"accelerator":  "1*nvidia-tesla-v100",
		"runasuser":    "1300",   // security context for fs access
		"runasgroup":   "1300",   // security context for fs access
		"fsgroup":      "1300",   // security context for fs access
		"pullpolicy":   "always", // pull new image always
		"privileged":   "true",
	}
	return jobTemplate
}

func CreateBarkInferenceScript(seed int, outputFileName string, prompts []string) string {
	promptReplacement := "prompts = [" + "\n"
	for _, v := range prompts {
		promptReplacement += "\"\"\"" + v + "\"\"\",\n"
	}
	promptReplacement += "]"

	script := strings.Replace(bark, "#$$$PROMPTS$$$", promptReplacement, 1)
	script = strings.Replace(script, "#$$$SEEDS$$$", fmt.Sprintf("set_seed(%d)", seed), 1)
	script = strings.Replace(script, "#$$$OUTPUTFILENAME$$$", outputFileName, 1)

	return script
}

func ExamplePrompts() []string {

	prompts := []string{`
We are in midst of a crowd of cows. This is a demonstration of a very long text which gets
converted to noise by Bark. Boy, so many cows. I am not sure if I can handle this.`,
		`Since the prompts can not be very long we split the text
into multiple prompts. The prompts get chained. Oh my god - the cows are still there! Run!
[music]`}

	return prompts
}
