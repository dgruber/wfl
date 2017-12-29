package simpletracker

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
)

func validateJobTemplate(jt drmaa2interface.JobTemplate) (bool, error) {
	if jt.InputPath != "" {
		if jt.InputPath == jt.OutputPath {
			return false, errors.New("InputPath in job template must not be the same than OutputPath")
		}
		if jt.InputPath == jt.ErrorPath {
			return false, errors.New("InputPath in job template must not be the same than ErrorPath")
		}
	}

	return true, nil
}
