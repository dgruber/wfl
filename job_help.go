package wfl

import (
	"context"
	"errors"

	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
)

// mergeJobTemplateWithDefaultTemplate adds requests from _def_ into _req_
// if specified in req
//
// Note there is no "unset" convention yet for the job template. It will
// only take actually implemented (in drmaa2os job tracker) settings into
// account.
func mergeJobTemplateWithDefaultTemplate(req, def drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
	if req.JobCategory == "" {
		req.JobCategory = def.JobCategory
	}
	if req.InputPath == "" {
		req.InputPath = def.InputPath
	}
	if req.OutputPath == "" {
		req.OutputPath = def.OutputPath
	}
	if req.ErrorPath == "" {
		req.ErrorPath = def.ErrorPath
	}
	if req.AccountingID == "" {
		req.AccountingID = def.AccountingID
	}
	if req.JobName == "" {
		req.JobName = def.JobName
	}
	if req.WorkingDirectory == "" {
		req.WorkingDirectory = def.WorkingDirectory
	}
	// replaces destination machines
	if req.CandidateMachines == nil && def.CandidateMachines != nil {
		if cm, err := copystructure.Copy(def.CandidateMachines); err == nil {
			req.CandidateMachines = cm.([]string)
		}
	}
	// replace extensions
	if req.ExtensionList == nil && def.ExtensionList != nil {
		if el, err := copystructure.Copy(def.ExtensionList); err == nil {
			req.ExtensionList = el.(map[string]string)
		}
	}
	// join files to stage
	req.StageInFiles = mergeStringMap(req.StageInFiles, def.StageInFiles)
	// join enviroment variables
	req.JobEnvironment = mergeStringMap(req.JobEnvironment, def.JobEnvironment)
	// TODO implement more when required
	return req
}

func mergeStringMap(dst, src map[string]string) map[string]string {
	if src != nil {
		if dst == nil {
			dst = make(map[string]string, len(src))
		}
		for k, v := range src {
			if dst[k] == "" {
				dst[k] = v
			}
		}
	}
	return dst
}

func waitForJobEndAndState(j *Job) drmaa2interface.JobState {
	job, jobArray, err := j.jobCheck()
	if err != nil {
		return drmaa2interface.Undetermined
	}
	if job != nil {
		lastError := job.WaitTerminated(drmaa2interface.InfiniteTime)
		if lastError != nil {
			return drmaa2interface.Undetermined
		}
		return job.GetState()
	}
	return jobArrayState(jobArray, true)
}

func jobArrayState(jobArray drmaa2interface.ArrayJob, wait bool) drmaa2interface.JobState {
	// it is a job array - waiting for each single task
	// if one of the tasks failed - the whole job array failed
	// if one of the tasks is undetermined and the rest is done, the array
	// is undetermined.
	jobArrayState := drmaa2interface.Done
	for _, job := range jobArray.GetJobs() {
		if wait {
			lastError := job.WaitTerminated(drmaa2interface.InfiniteTime)
			if lastError != nil {
				return drmaa2interface.Undetermined
			}
		}
		switch job.GetState() {
		case drmaa2interface.Done:
			continue
		case drmaa2interface.Failed:
			// overwrites all
			jobArrayState = drmaa2interface.Failed
		case drmaa2interface.Undetermined:
			// overwrites done
			if jobArrayState == drmaa2interface.Done {
				jobArrayState = drmaa2interface.Undetermined
			}
		}
	}
	return jobArrayState
}

func waitArrayJobTerminated(jobArray drmaa2interface.ArrayJob) error {
	var lastErr error
	for _, job := range jobArray.GetJobs() {
		err := job.WaitTerminated(drmaa2interface.InfiniteTime)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (j *Job) lastJob() *task {
	if len(j.tasklist) == 0 {
		return nil
	}
	return j.tasklist[len(j.tasklist)-1]
}

func (j *Job) jobCheck() (drmaa2interface.Job, drmaa2interface.ArrayJob, error) {
	task := j.lastJob()
	if task == nil {
		j.errorf(j.ctx, "jobCheck(): task is nil")
		return nil, nil, errors.New("job task not available")
	} else if task.job == nil && task.jobArray == nil {
		j.errorf(j.ctx, "jobCheck(): task has no drmaa2 job")
		return nil, nil, errors.New("job not available")
	}
	return task.job, task.jobArray, nil
}

func (j *Job) checkCtx() error {
	if j.wfl == nil {
		return errors.New("no workflow defined")
	}
	if j.wfl.ctx == nil {
		return errors.New("no context defined")
	}
	return nil
}

func (j *Job) begin(ctx context.Context, f string) {
	if j == nil || j.wfl == nil || j.wfl.log == nil {
		return
	}
	j.wfl.log.Begin(ctx, f)
}

func (j *Job) infof(ctx context.Context, s string, args ...interface{}) {
	if j == nil || j.wfl == nil || j.wfl.log == nil {
		return
	}
	j.wfl.log.Infof(ctx, s, args...)
}
func (j *Job) warningf(ctx context.Context, s string, args ...interface{}) {
	if j == nil || j.wfl == nil || j.wfl.log == nil {
		return
	}
	j.wfl.log.Warningf(ctx, s, args...)
}

func (j *Job) errorf(ctx context.Context, s string, args ...interface{}) {
	if j == nil || j.wfl == nil || j.wfl.log == nil {
		return
	}
	j.wfl.log.Errorf(ctx, s, args...)
}
