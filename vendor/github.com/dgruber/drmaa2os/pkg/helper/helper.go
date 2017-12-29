package helper

import (
	"encoding/json"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
)

func ArrayJobID2GUIDs(id string) ([]string, error) {
	var guids []string
	err := json.Unmarshal([]byte(id), &guids)
	if err != nil {
		return nil, err
	}
	return guids, nil
}

func Guids2ArrayJobID(guids []string) string {
	id, err := json.Marshal(guids)
	if err != nil {
		return ""
	}
	return string(id)
}

func AddArrayJobAsSingleJobs(jt drmaa2interface.JobTemplate, t jobtracker.JobTracker, begin int, end int, step int) (string, error) {
	var guids []string
	for i := begin; i <= end; i += step {
		guid, err := t.AddJob(jt)
		if err != nil {
			return Guids2ArrayJobID(guids), err
		}
		guids = append(guids, guid)
	}
	return Guids2ArrayJobID(guids), nil
}
