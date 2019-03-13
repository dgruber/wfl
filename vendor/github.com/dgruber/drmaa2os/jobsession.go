package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/d2hlp"
	"github.com/dgruber/drmaa2os/pkg/jobtracker"
	"time"
)

// JobSession instance acts as container for job instances controlled
// through the DRMAA API. The session methods support the submission
// of new jobs and the monitoring of existing jobs.
type JobSession struct {
	name    string
	tracker []jobtracker.JobTracker
}

// Close MUST perform the necessary action to disengage from the DRM system.
// It SHOULD be callable only once, by only one of the application threads.
// This SHOULD be ensured by the library implementation. Additional calls
// beyond the first one SHOULD lead to a InvalidSessionException error
// notification.
// The corresponding state information MUST be saved to some stable storage
// before the method returns. This method SHALL NOT affect any jobs or
// reservations in the session (e.g., queued and running jobs remain queued
// and running). (TODO)
func (js *JobSession) Close() error {
	if js.name == "" && js.tracker == nil {
		return ErrorInvalidSession
	}
	js.name = ""
	js.tracker = nil
	return nil
}

// GetContact method reports the contact value that was used in the
// SessionManager::createJobSession call for this instance. If no
// value was originally provided, the default contact string from the
// implementation MUST be returned.
func (js *JobSession) GetContact() (string, error) {
	return "", nil
}

// GetSessionName reports the session name, a value that resulted from the
// SessionManager::createJobSession or SessionManager::openJobSession
// call for this instance.
func (js *JobSession) GetSessionName() (string, error) {
	return js.name, nil
}

// GetJobCategories provides the list of valid job category names which
// can be used for the jobCategory attribute in a JobTemplate instance.
func (js *JobSession) GetJobCategories() ([]string, error) {
	var lastError error
	var jobCategories []string
	for _, tracker := range js.tracker {
		cat, err := tracker.ListJobCategories()
		if err != nil {
			lastError = err
			continue
		}
		jobCategories = append(jobCategories, cat...)
	}
	return jobCategories, lastError
}

// GetJobs returns the set of jobs that belong to the job session. The
// filter parameter allows to choose a subset of the session jobs as
// return value. If no job matches or the session has no jobs attached,
// the method MUST return an empty set. If filter is UNSET, all session
// jobs MUST be returned.
// Time-dependent effects of this method, such as jobs no longer matching
// to filter criteria on evaluation time, are implementation-specific.
// The purpose of the filter parameter is to keep scalability with a
// large number of jobs per session. Applications therefore must consider
// the possibly changed state of jobs during their evaluation of the method
// result.
func (js *JobSession) GetJobs(filter drmaa2interface.JobInfo) ([]drmaa2interface.Job, error) {
	var joblist []drmaa2interface.Job

	for _, tracker := range js.tracker {
		jobs, err := tracker.ListJobs()
		if err != nil {
			return nil, err
		}
		for _, jobid := range jobs {
			if jinfo, err := tracker.JobInfo(jobid); err != nil {
				continue
			} else {
				if d2hlp.JobInfoMatches(jinfo, filter) {
					// TODO get template from DB
					jobtemplate := drmaa2interface.JobTemplate{}

					job := newJob(jobid, js.name, jobtemplate, tracker)
					joblist = append(joblist, drmaa2interface.Job(job))
				}
			}
		}
	}

	return joblist, nil
}

// GetJobArray method returns the JobArray instance with the given ID.
// If the session does not / no longer contain the according job array,
// InvalidArgumentException SHALL be thrown.
func (js *JobSession) GetJobArray(id string) (drmaa2interface.ArrayJob, error) {
	var joblist []drmaa2interface.Job
	for _, tracker := range js.tracker {
		jobids, err := tracker.ListArrayJobs(id)
		if err != nil {
			return nil, err
		}
		for _, id := range jobids {
			// TODO get template from DB
			jobtemplate := drmaa2interface.JobTemplate{}

			job := newJob(id, js.name, jobtemplate, tracker)
			joblist = append(joblist, drmaa2interface.Job(job))
		}
	}
	return newArrayJob(id, js.name, drmaa2interface.JobTemplate{}, joblist), nil
}

// RunJob method submits a job with the attributes defined in the given job template
// instance. The method returns a Job object that represents the job in the underlying
// DRM system. Depending on the job template settings, submission attempts may be
// rejected with an InvalidArgumentException. The error details SHOULD provide further
// information about the attribute(s) responsible for the rejection. When this method
// returns a valid Job instance, the following conditions SHOULD be fulfilled:
// - The job is part of the persistent state of the job session.
// - All non-DRMAA and DRMAA interfaces to the DRM system report the job as
//   being submitted to the DRM system.
// - The job has one of the DRMAA job states.
func (js *JobSession) RunJob(jt drmaa2interface.JobTemplate) (drmaa2interface.Job, error) {
	id, err := js.tracker[0].AddJob(jt)
	if err != nil {
		return nil, err
	}
	return newJob(id, js.name, jt, js.tracker[0]), nil
}

// RunBulkJobs method creates a set of parametric jobs, each with attributes as defined
// in the given job template instance.
func (js *JobSession) RunBulkJobs(jt drmaa2interface.JobTemplate, begin, end, step, maxParallel int) (drmaa2interface.ArrayJob, error) {
	id, err := js.tracker[0].AddArrayJob(jt, begin, end, step, maxParallel)
	if err != nil {
		return nil, err
	}
	return js.GetJobArray(id)
}

// WaitAnyStarted method blocks until any of the jobs referenced in the jobs
// parameter entered one of the "Started" states.
//
// The timeout argument specifies the desired waiting time for the state change.
// The constant value drmaa2interface.InfiniteTime MUST be supported to get an
// indefinite waiting time. The constant value drmaa2interface.ZeroTime MUST be
// supported to express that the method call SHALL return immediately.
// A time.Duration can be specified to indicate the maximum waiting time.
// If the method call returns because of timeout, an TimeoutException SHALL be
// raised.
func (js *JobSession) WaitAnyStarted(jobs []drmaa2interface.Job, timeout time.Duration) (drmaa2interface.Job, error) {
	return waitAny(true, jobs, timeout)
}

// WaitAnyTerminated method blocks until any of the jobs referenced in the
// jobs parameter entered one of the "Terminated" states.
//
// The timeout argument specifies the desired waiting time for the state change.
// The constant value drmaa2interface.InfiniteTime MUST be supported to get an
// indefinite waiting time. The constant value drmaa2interface.ZeroTime MUST be
// supported to express that the method call SHALL return immediately.
// A time.Duration can be specified to indicate the maximum waiting time.
// If the method call returns because of timeout, an TimeoutException SHALL be
// raised.
func (js *JobSession) WaitAnyTerminated(jobs []drmaa2interface.Job, timeout time.Duration) (drmaa2interface.Job, error) {
	return waitAny(false, jobs, timeout)
}
