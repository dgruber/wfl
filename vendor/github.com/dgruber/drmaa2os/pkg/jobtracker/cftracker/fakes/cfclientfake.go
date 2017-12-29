package fake

import (
	"errors"
	"github.com/cloudfoundry-community/go-cfclient"
	"time"
)

func createFailedFakeTask() (t cfclient.Task) {
	t.GUID = "GUID"
	t.SequenceID = 1
	t.Name = "name"
	t.Command = "command"
	t.State = "FAILED"
	t.MemoryInMb = 1024
	t.DiskInMb = 1024
	// t.Result =  {FailureReason = ""}
	t.CreatedAt = time.Date(2016, 12, 22, 13, 24, 20, 0, time.FixedZone("UTC", 0))
	t.UpdatedAt = time.Date(2016, 12, 23, 13, 24, 20, 0, time.FixedZone("UTC", 0))
	t.DropletGUID = "dropletGUID"
	return t
}

type cfclientfake struct {
}

func NewClientFake() *cfclientfake {
	cf := cfclientfake{}
	return &cf
}

func (cf *cfclientfake) ListTasks() ([]cfclient.Task, error) {
	tasks := make([]cfclient.Task, 0, 1)
	tasks = append(tasks, createFailedFakeTask())
	return tasks, nil
}

func (cf *cfclientfake) CreateTask(tr cfclient.TaskRequest) (t cfclient.Task, err error) {
	if tr.Command == "error" {
		return t, errors.New("error")
	}
	return createFailedFakeTask(), nil
}

func (cf *cfclientfake) TaskByGuid(task string) (t cfclient.Task, err error) {
	if task == "error" {
		return t, errors.New("error")
	}
	t = createFailedFakeTask()
	switch task {
	case "PENDING":
		t.State = "PENDING"
	case "RUNNING":
		t.State = "RUNNING"
	case "CANCELING":
		t.State = "CANCELING"
	case "SUCCEEDED":
		t.State = "SUCCEEDED"
	case "FAILED":
		t.State = "FAILED"
	case "unknown":
		t.State = "unknown"
	}
	return t, err
}

func (cf *cfclientfake) TerminateTask(task string) error {
	if task == "noerror" {
		return nil
	}
	return errors.New("error")
}

func (cf *cfclientfake) ListApps() ([]cfclient.App, error) {
	apps := make([]cfclient.App, 0, 1)
	apps = append(apps, cfclient.App{Name: "name", Guid: "guid"})
	return apps, nil
}
