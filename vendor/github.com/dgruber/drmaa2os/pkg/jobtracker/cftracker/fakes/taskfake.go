package fake

import (
	"github.com/cloudfoundry-community/go-cfclient"
	_ "github.com/dgruber/drmaa2interface"
	"time"
)

func FailedTaskFake() *cfclient.Task {
	return &cfclient.Task{
		GUID:    "123",
		Name:    "name",
		State:   "FAILED",
		Command: "command",
		// Result
		SequenceID:  1,
		MemoryInMb:  1024,
		DiskInMb:    2048,
		CreatedAt:   time.Date(2016, 12, 22, 13, 24, 20, 0, time.FixedZone("UTC", 0)),
		UpdatedAt:   time.Date(2017, 12, 22, 13, 24, 20, 0, time.FixedZone("UTC", 0)),
		DropletGUID: "dropletGUID",
	}
}

func SucceededTaskFake() *cfclient.Task {
	return &cfclient.Task{
		GUID:    "123",
		Name:    "name",
		State:   "SUCCEEDED",
		Command: "command",
		// Result
		SequenceID:  1,
		MemoryInMb:  1024,
		DiskInMb:    2048,
		CreatedAt:   time.Date(2016, 12, 22, 13, 24, 20, 0, time.FixedZone("UTC", 0)),
		UpdatedAt:   time.Date(2017, 12, 22, 13, 24, 20, 0, time.FixedZone("UTC", 0)),
		DropletGUID: "dropletGUID",
	}
}

func TaskRequestFake() *cfclient.TaskRequest {
	return &cfclient.TaskRequest{}
}
