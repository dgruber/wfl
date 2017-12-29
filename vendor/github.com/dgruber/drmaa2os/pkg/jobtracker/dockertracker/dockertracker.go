package dockertracker

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/helper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"golang.org/x/net/context"
	"io"
	"os"
	"strings"
	"time"
)

type DockerTracker struct {
	cli *client.Client
}

// New creates a new DockerTracker. How the Docker client
// is configured can be influenced by (from the Docker
// Documentation (https://github.com/moby/moby/blob/master/client/client.go)):
// "Use DOCKER_HOST to set the url to the docker server.
// Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
// Use DOCKER_CERT_PATH to load the TLS certificates from.
// Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default."
func New() (*DockerTracker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	_, errPing := cli.Ping(context.Background())
	if errPing != nil {
		return nil, err
	}
	return &DockerTracker{cli: cli}, nil
}

func (dt *DockerTracker) ListJobs() ([]string, error) {
	if err := dt.check(); err != nil {
		return nil, err
	}
	containers, err := dt.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	return containersToJobList(containers), nil
}

func (dt *DockerTracker) AddJob(jt drmaa2interface.JobTemplate) (string, error) {
	if err := dt.check(); err != nil {
		return "", err
	}

	if err := checkJobTemplate(jt); err != nil {
		return "", err
	}

	// stage image

	// https://docs.docker.com/engine/api/#api-example

	// https://github.com/moby/moby/blob/master/api/types/container/config.go
	config, err := jobTemplateToContainerConfig(jt)
	if err != nil {
		return "", err
	}

	hostConfig, err := jobTemplateToHostConfig(jt)
	if err != nil {
		return "", fmt.Errorf("Docker Host Config: %s", err.Error())
	}

	networkingConfig, err := jobTemplateToNetworkingConfig(jt)
	if err != nil {
		return "", fmt.Errorf("Docker Network Config: %s", err.Error())
	}

	// pull image -> requires internet access
	//_, err = dt.cli.ImagePull(context.Background(), jt.JobCategory, types.ImagePullOptions{})
	// if err != nil {
	//	return "", fmt.Errorf("Error while pulling image: %s", err.Error())
	// }
	ccBody, err := dt.cli.ContainerCreate(context.Background(),
		config,
		hostConfig,
		networkingConfig,
		jt.JobName)

	if err != nil {
		return "", fmt.Errorf("Error while creating container: %s", err.Error())
	}

	err = dt.cli.ContainerStart(context.Background(), ccBody.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("Error while starting container: %s", err.Error())
	}

	if jt.OutputPath != "" || jt.ErrorPath != "" {
		stdout := false
		stderr := false

		if jt.OutputPath != "" {
			stdout = true
		}
		if jt.ErrorPath != "" {
			stderr = true
		}

		handleInputOutput(dt.cli,
			ccBody.ID,
			types.ContainerAttachOptions{Stream: true, Stdout: stdout, Stderr: stderr, Logs: true},
			jt.OutputPath,
			jt.ErrorPath)
	}
	return ccBody.ID, nil
}

func handleInputOutput(cli *client.Client, id string, options types.ContainerAttachOptions, stdoutfile, stderrfile string) {
	res, err := cli.ContainerAttach(context.Background(), id, options)
	if err != nil {
		panic(err)
	}
	if stdoutfile != "" && stderrfile != "" {
		redirectOut(res, stdoutfile, stderrfile)
	} else if stdoutfile != "" {
		redirect(res, stdoutfile)
	} else if stderrfile != "" {
		redirect(res, stderrfile)
	}
}

func redirectOut(res types.HijackedResponse, outfilename, errfilename string) {
	go func() {
		outfile, err := os.Create(outfilename)
		if err != nil {
			panic(err)
		}
		errfile, err := os.Create(errfilename)
		if err != nil {
			panic(err)
		}

		stdcopy.StdCopy(outfile, errfile, res.Reader)
		outfile.Close()
		errfile.Close()
		res.Close()
	}()
}

func redirect(res types.HijackedResponse, file string) {
	go func() {
		buf := make([]byte, 1)
		file, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		io.CopyBuffer(file, res.Reader, buf)
		file.Close()
		res.Close()
	}()
}

func (dt *DockerTracker) AddArrayJob(jt drmaa2interface.JobTemplate, begin int, end int, step int, maxParallel int) (string, error) {
	return helper.AddArrayJobAsSingleJobs(jt, dt, begin, end, step)

}

func (dt *DockerTracker) ListArrayJobs(id string) ([]string, error) {
	if err := dt.check(); err != nil {
		return nil, err
	}
	return helper.ArrayJobID2GUIDs(id)

}

func (dt *DockerTracker) JobState(jobid string) drmaa2interface.JobState {
	if err := dt.check(); err != nil {
		return drmaa2interface.Undetermined
	}
	container, err := dt.cli.ContainerInspect(context.Background(), jobid)
	if err != nil {
		return drmaa2interface.Undetermined
	}
	if container.State == nil {
		return drmaa2interface.Undetermined
	}
	return containerToDRMAA2State(container.State)
}

func (dt *DockerTracker) JobInfo(jobid string) (ji drmaa2interface.JobInfo, err error) {
	if err := dt.check(); err != nil {
		return ji, err
	}
	container, err := dt.cli.ContainerInspect(context.Background(), jobid)
	if err != nil {
		return ji, err
	}
	return containerToDRMAA2JobInfo(container)
}

func (dt *DockerTracker) JobControl(jobid, state string) error {
	if err := dt.check(); err != nil {
		return err
	}
	switch state {
	case "suspend":
		return dt.cli.ContainerKill(context.Background(), jobid, "SIGSTOP")
	case "resume":
		return dt.cli.ContainerKill(context.Background(), jobid, "SIGCONT")
	case "hold":
		return errors.New("Unsupported Operation")
	case "release":
		return errors.New("Unsupported Operation")
	case "terminate":
		return dt.cli.ContainerKill(context.Background(), jobid, "SIGKILL")
	}
	return errors.New("undefined state")
}

func (dt *DockerTracker) Wait(jobid string, timeout time.Duration, states ...drmaa2interface.JobState) error {
	// ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel()

	// dt.cli.ContainerWait(ctx, jobid, "")
	//  moby/api/types/container/waitcondition.go:
	// "Possible WaitCondition Values.
	//
	// WaitConditionNotRunning (default) is used to wait for any of the non-running
	// states: "created", "exited", "dead", "removing", or "removed".
	//
	// WaitConditionNextExit is used to wait for the next time the state changes
	// to a non-running state. If the state is currently "created" or "exited",
	// this would cause Wait() to block until either the container runs and exits
	// or is removed.
	//
	// WaitConditionRemoved is used to wait for the container to be removed."

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	hasStateCh := make(chan bool, 1)
	defer close(hasStateCh)

	quit := make(chan bool)

	go func() {
		t := time.NewTicker(time.Millisecond * 200)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				currentState := dt.JobState(jobid)
				if isInExpectedState(currentState, states...) {
					hasStateCh <- true
					return
				}
			case <-quit:
				return
			}
		}
	}()

	select {
	case <-ticker.C:
		quit <- true
		return errors.New("timeout while waiting for job state")
	case <-hasStateCh:
		return nil
	}
	return nil
}

// DeleteJob removes a container so it is no longer in docker ps -a (and therefore not in the job list).
func (dt *DockerTracker) DeleteJob(jobid string) error {
	if err := dt.check(); err != nil {
		return err
	}
	if state := dt.JobState(jobid); state != drmaa2interface.Done && state != drmaa2interface.Failed {
		return errors.New("job is not in an end-state")
	}
	return dt.cli.ContainerRemove(context.Background(),
		jobid,
		types.ContainerRemoveOptions{
			Force:         true,
			RemoveLinks:   false,
			RemoveVolumes: true,
		},
	)
}

// ListJobCategories lists all containers available to run commands on.
func (dt *DockerTracker) ListJobCategories() ([]string, error) {
	if err := dt.check(); err != nil {
		return nil, err
	}
	images, err := dt.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(images))
	for _, i := range images {
		ids = append(ids, strings.Join(i.RepoTags, "/"))
	}
	return ids, nil
}

func (dt *DockerTracker) check() error {
	if dt == nil || dt.cli == nil {
		return errors.New("DockerTracker not initialized")
	}
	return nil
}
