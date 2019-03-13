package dockertracker

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/helper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"golang.org/x/net/context"
	"strings"
	"time"
)

type DockerTracker struct {
	jobsession string
	cli        *client.Client
}

// New creates a new DockerTracker. How the Docker client
// is configured can be influenced by (from the Docker
// Documentation (https://github.com/moby/moby/blob/master/client/client.go)):
// "Use DOCKER_HOST to set the url to the docker server.
//  Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
//  Use DOCKER_CERT_PATH to load the TLS certificates from.
//  Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default."
func New(jobsession string) (*DockerTracker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	_, errPing := cli.Ping(context.Background())
	if errPing != nil {
		return nil, err
	}
	return &DockerTracker{cli: cli, jobsession: jobsession}, nil
}

func (dt *DockerTracker) ListJobs() ([]string, error) {
	if err := dt.check(); err != nil {
		return nil, err
	}
	f := filters.NewArgs()
	f.Add("label", "drmaa2_jobsession="+dt.jobsession)
	containers, err := dt.cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: f, All: true})
	if err != nil {
		return nil, err
	}
	return containersToJobList(dt.jobsession, containers), nil
}

func (dt *DockerTracker) AddJob(jt drmaa2interface.JobTemplate) (string, error) {
	if err := dt.check(); err != nil {
		return "", err
	}
	return runJob(dt.jobsession, dt.cli, jt)
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
	// add:
	// stats, err := dt.cli.ContainerStats(context.Background(), jobid, false)
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

	return helper.WaitForState(dt, jobid, timeout, states...)
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

// ListJobCategories lists all container images available to run commands on.
func (dt *DockerTracker) ListJobCategories() ([]string, error) {
	if err := dt.check(); err != nil {
		return nil, err
	}
	images, err := dt.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(images))
	for _, i := range images {
		if len(i.RepoTags) > 0 {
			ids = append(ids, strings.Join(i.RepoTags, "/"))
		}
	}
	return ids, nil
}

func (dt *DockerTracker) check() error {
	if dt == nil || dt.cli == nil {
		return errors.New("DockerTracker not initialized")
	}
	return nil
}
