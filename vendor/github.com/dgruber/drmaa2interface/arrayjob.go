package drmaa2interface

// ArrayJob defines all methods which can be executed on a
// DRMAA2 job array. A job array contains a set of jobs with
// the same properties (same job template).
type ArrayJob interface {
	GetID() string
	GetJobs() []Job
	GetSessionName() string
	GetJobTemplate() JobTemplate
	Suspend() error
	Resume() error
	Hold() error
	Release() error
	Terminate() error
}
