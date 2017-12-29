package drmaa2os

import (
	"github.com/dgruber/drmaa2interface"
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker"
)

type ArrayJob struct {
	id          string
	sessionname string
	template    drmaa2interface.JobTemplate
	jobs        []drmaa2interface.Job
}

func NewArrayJob(id, jsessionname string, tmpl drmaa2interface.JobTemplate, jobs []drmaa2interface.Job) *ArrayJob {
	return &ArrayJob{
		id:          id,
		sessionname: jsessionname,
		template:    tmpl,
		jobs:        jobs,
	}
}

func (aj *ArrayJob) GetID() string {
	return aj.id
}

func (aj *ArrayJob) GetJobs() []drmaa2interface.Job {
	return aj.jobs
}

func (aj *ArrayJob) GetSessionName() string {
	return aj.sessionname
}

func (aj *ArrayJob) GetJobTemplate() drmaa2interface.JobTemplate {
	return aj.template
}

func (aj *ArrayJob) Suspend() error {
	return jobAction(suspend, aj.jobs)
}

func (aj *ArrayJob) Resume() error {
	return jobAction(resume, aj.jobs)
}

func (aj *ArrayJob) Hold() error {
	return jobAction(hold, aj.jobs)
}

func (aj *ArrayJob) Release() error {
	return jobAction(release, aj.jobs)
}

func (aj *ArrayJob) Terminate() error {
	return jobAction(terminate, aj.jobs)
}
