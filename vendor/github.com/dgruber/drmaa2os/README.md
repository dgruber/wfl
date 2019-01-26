# drmaa2os - A Go API for OS Processes, Docker Containers, Cloud Foundry Tasks, Kubernetes Jobs, Grid Engine Jobs and more...

_DRMAA2 for OS processes and more_

[![CircleCI](https://circleci.com/gh/dgruber/drmaa2os.svg?style=svg)](https://circleci.com/gh/dgruber/drmaa2os)
[![codecov](https://codecov.io/gh/dgruber/drmaa2os/branch/master/graph/badge.svg)](https://codecov.io/gh/dgruber/drmaa2os)

This is a Go API based on an open standard ([Open Grid Forum DRMAA2](https://www.ogf.org/documents/GFD.231.pdf)) for submitting and
supervising workloads running in operating system processes, containers, PODs, tasks, or HPC batch jobs.

The API allows you to develop and run job workflows in OS processes and switch later to 
containers running in Kubernetes, as Cloud Foundry tasks, pure Docker, or Singularity
without changing the application logic.

Its main pupose is supporting you with an abstraction layer on top of platforms, workload managers, 
and HPC cluster schedulers, so that you don't need to deal with the underlaying details and differences.

An even simpler interface for creating job workflows without dealing with the DRMAA2 details is
[*wfl*](https://github.com/dgruber/wfl) which is based on the Go DRMAA2 implementation.

For details about the mapping of job operations please consult the platform specific READMEs:

  * [OS Processes](pkg/jobtracker/simpletracker/README.md)
  * [Cloud Foundry](pkg/jobtracker/cftracker/README.md)
  * [Docker / Moby](pkg/jobtracker/dockertracker/README.md)
  * [Kubernetes](pkg/jobtracker/kubernetestracker/README.md)
  * [Singularity](pkg/jobtracker/singularity/README.md)

[Feedback](mailto:info@gridengine.eu) welcome!

For a Go DRMAA2 wrapper based on C DRMAA2 (_libdrmaa2.so_) like for *Univa Grid Engine* please
check out [drmaa2](https://github.com/dgruber/drmaa2).

## Basic Usage

Following example demonstrates how a job running as OS process can be executed. More examples can be found in the _examples_ subdirectory. 

```go
	sm, _ := drmaa2os.NewDefaultSessionManager("testdb.db")

	js, _ := sm.CreateJobSession("jobsession", "")

	jt := drmaa2interface.JobTemplate{
		JobName:       "job1",
		RemoteCommand: "sleep",
		Args:          []string{"2"},
	}

	job, _ := js.RunJob(jt)

	job.WaitTerminated(drmaa2interface.InfiniteTime)

	if job.GetState() == drmaa2interface.Done {
		job2, _ := js.RunJob(jt)
		job2.WaitTerminated(drmaa2interface.InfiniteTime)
	} else {
		fmt.Println("Failed to execute job1 successfully")
	}

	js.Close()
	sm.DestroyJobSession("jobsession")
```

## Using other Backends

### Docker

### Kubernetes

### Cloud Foundry

### Singularity





