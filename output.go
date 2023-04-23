package wfl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dgruber/drmaa2interface"
)

func getJobOutputKubernetes(job drmaa2interface.Job) (string, error) {
	ji, err := job.GetJobInfo()
	if err != nil {
		return "", fmt.Errorf("failed getting job info: %s", err)
	}
	if ji.ExtensionList != nil {
		if output, ok := ji.ExtensionList["output"]; ok {
			if len(output) > 0 && output[len(output)-1] == '\n' {
				output = output[:len(output)-1]
			}
			if len(output) > 0 && output[len(output)-1] == '\r' {
				output = output[:len(output)-1]
			}
			return output, nil
		}
	}
	return "", errors.New("no output in jobinfo")
}

func getJobOutputOS(job drmaa2interface.Job) (string, error) {
	template, err := job.GetJobTemplate()
	if err != nil {
		return "", fmt.Errorf("failed getting job template: %s", err)
	}
	if !isPathLocalFile(template.OutputPath) {
		return "", fmt.Errorf("output path %s is not a local file",
			template.OutputPath)
	}
	return getFileContent(template.OutputPath)
}

func getJobOutputDocker(job drmaa2interface.Job) (string, error) {
	template, err := job.GetJobTemplate()
	if err != nil {
		return "", fmt.Errorf("failed getting job template: %s", err)
	}
	if !isPathLocalFile(template.OutputPath) {
		return "", fmt.Errorf("output path %s is not a local file",
			template.OutputPath)
	}
	return getFileContent(template.OutputPath)
}

func getFileContent(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed reading file %s: %s",
			path, err)
	}

	// OS process -> the output is in the file
	// remove trailing newline cr
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data = data[:len(data)-1]
	}
	return string(data), nil
}

func isPathLocalFile(path string) bool {
	normalFile := path != "/dev/null" &&
		path != "/dev/stdout" &&
		path != "/dev/stderr" &&
		path != ""

	if normalFile {
		// check if path is a file which exists with os.Stat
		fi, err := os.Stat(path)
		if err != nil {
			return false
		}
		return !fi.IsDir()
	}
	return false
}

func getJobOutpuForJob(wflType SessionManagerType, job drmaa2interface.Job) (string, error) {

	state := job.GetState()
	if state == drmaa2interface.Undetermined {
		return "", errors.New("job state is undetermined")
	}

	err := job.WaitTerminated(drmaa2interface.InfiniteTime)
	if err != nil {
		return "", fmt.Errorf("failed waiting for job termination: %s", err)
	}

	// for Kubernetes we need the jobinfo "output" extension
	switch wflType {

	case KubernetesSessionManager:
		{
			output, err := getJobOutputKubernetes(job)
			if err != nil {
				return "", fmt.Errorf("failed getting job info for k8s job %s: %s",
					job.GetID(), err)
			}
			return output, nil
		}
	case DockerSessionManager:
		{
			output, err := getJobOutputDocker(job)
			if err != nil {
				return "", fmt.Errorf("failed getting job info for docker job %s: %s",
					job.GetID(), err)
			}
			return output, nil
		}
	case DefaultSessionManager:
		{
			output, err := getJobOutputOS(job)
			if err != nil {
				return "", fmt.Errorf("failed getting job info for OS job %s: %s",
					job.GetID(), err)
			}
			return output, nil
		}
	}

	return "", errors.New("unsupported workflow type")
}
