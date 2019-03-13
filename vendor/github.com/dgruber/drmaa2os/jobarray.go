package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
)

// ArrayJob represents a set of jobs created by one operation. In DRMAA,
// JobArray instances are only created by the RunBulkJobs method.
// JobArray instances differ from the JobList data structure due to their
// potential for representing a DRM system concept, while JobList
// is a DRMAA-only concept realized by language binding support.
type ArrayJob struct {
	id          string
	sessionname string
	template    drmaa2interface.JobTemplate
	jobs        []drmaa2interface.Job
}

// GetID reports the job identifier assigned to the job array by the DRM
// system in text form.
func (aj *ArrayJob) GetID() string {
	return aj.id
}

// GetJobs provides the list of jobs that are part of the job array, regardless
// of their state.
func (aj *ArrayJob) GetJobs() []drmaa2interface.Job {
	return aj.jobs
}

// GetSessionName states the name of the JobSession that was used to create the
// bulk job represented by this instance. If the session name cannot be determined,
// for example since the bulk job was created outside of a DRMAA session, the
// attribute SHOULD have an UNSET value (i.e. is "").
func (aj *ArrayJob) GetSessionName() string {
	return aj.sessionname
}

// GetJobTemplate provides a reference to a JobTemplate instance that has equal
// values to the one that was used for the job submission creating this JobArray
// instance.
func (aj *ArrayJob) GetJobTemplate() drmaa2interface.JobTemplate {
	return aj.template
}

// Suspend triggers a job state transition from RUNNING to SUSPENDED state.
//
// The job control functions allow modifying the status of the job array in the DRM system,
// with the same semantic as in the Job interface. If one of the jobs in the array is in
// an inappropriate state for the particular method, the method MAY raise an
// InvalidStateException.
//
// The methods SHOULD return after the action has been acknowledged by the DRM system for
// all jobs in the array, but MAY return before the action has been completed for all of
// the jobs. Some  DRMAA implementations MAY allow this method to be used to control job
// arrays created externally to the DRMAA session. This behavior is implementation-specific.
func (aj *ArrayJob) Suspend() error {
	return jobAction(suspend, aj.jobs)
}

// Resume triggers a job state transition from SUSPENDED to RUNNING state.
//
// The job control functions allow modifying the status of the job array in the DRM system,
// with the same semantic as in the Job interface. If one of the jobs in the array is in
// an inappropriate state for the particular method, the method MAY raise an
// InvalidStateException.
//
// The methods SHOULD return after the action has been acknowledged by the DRM system for
// all jobs in the array, but MAY return before the action has been completed for all of
// the jobs. Some  DRMAA implementations MAY allow this method to be used to control job
// arrays created externally to the DRMAA session. This behavior is implementation-specific.
func (aj *ArrayJob) Resume() error {
	return jobAction(resume, aj.jobs)
}

// Hold triggers a transition from QUEUED to QUEUED_HELD, or from REQUEUED to
// REQUEUED_HELD state.
//
// The job control functions allow modifying the status of the job array in the DRM system,
// with the same semantic as in the Job interface. If one of the jobs in the array is in
// an inappropriate state for the particular method, the method MAY raise an
// InvalidStateException.
//
// The methods SHOULD return after the action has been acknowledged by the DRM system for
// all jobs in the array, but MAY return before the action has been completed for all of
// the jobs. Some  DRMAA implementations MAY allow this method to be used to control job
// arrays created externally to the DRMAA session. This behavior is implementation-specific.
func (aj *ArrayJob) Hold() error {
	return jobAction(hold, aj.jobs)
}

// Release triggers a transition from QUEUED_HELD to QUEUED, or from REQUEUED_HELD
// to REQUEUED state.
//
// The job control functions allow modifying the status of the job array in the DRM system,
// with the same semantic as in the Job interface. If one of the jobs in the array is in
// an inappropriate state for the particular method, the method MAY raise an
// InvalidStateException.
//
// The methods SHOULD return after the action has been acknowledged by the DRM system for
// all jobs in the array, but MAY return before the action has been completed for all of
// the jobs. Some  DRMAA implementations MAY allow this method to be used to control job
// arrays created externally to the DRMAA session. This behavior is implementation-specific.
func (aj *ArrayJob) Release() error {
	return jobAction(release, aj.jobs)
}

// Terminate triggers a transition from any of the "Started" states to one of the
// "Terminated" states.
//
// The job control functions allow modifying the status of the job array in the DRM system,
// with the same semantic as in the Job interface. If one of the jobs in the array is in
// an inappropriate state for the particular method, the method MAY raise an
// InvalidStateException.
//
// The methods SHOULD return after the action has been acknowledged by the DRM system for
// all jobs in the array, but MAY return before the action has been completed for all of
// the jobs. Some  DRMAA implementations MAY allow this method to be used to control job
// arrays created externally to the DRMAA session. This behavior is implementation-specific.
func (aj *ArrayJob) Terminate() error {
	return jobAction(terminate, aj.jobs)
}
