package cftracker

import (
	"errors"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dgruber/drmaa2interface"
	"strings"
)

func convertTasksInNames(t []cfclient.Task) []string {
	if len(t) == 0 {
		return []string{}
	}
	names := make([]string, 0, len(t))
	for _, task := range t {
		names = append(names, task.GUID)
	}
	return names
}

func convertTaskInJobinfo(t cfclient.Task) (ji drmaa2interface.JobInfo) {
	switch t.State {
	case "FAILED":
		ji.State = drmaa2interface.Failed
	case "SUCCEEDED":
		ji.State = drmaa2interface.Done
	}
	return ji
}

func convertJobTemplateInTaskRequest(jt drmaa2interface.JobTemplate) (tr cfclient.TaskRequest, err error) {
	if jt.RemoteCommand == "" {
		return tr, errors.New("RemoteCommand is not set in JobTemplate")
	}
	tr.Command = jt.RemoteCommand
	if len(jt.Args) > 0 {
		tr.Command += " " + strings.Join(jt.Args, " ")
	}

	if jt.JobCategory == "" {
		return tr, errors.New("JobCategory is not set in JobTemplate")
	}
	tr.DropletGUID = jt.JobCategory

	tr.Name = jt.JobName
	if jt.MinPhysMemory > 0 {
		tr.MemoryInMegabyte = int(jt.MinPhysMemory/1024) + 1
	}
	return tr, err
}
