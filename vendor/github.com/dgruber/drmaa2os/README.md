# drmaa2os - A Go API for OS Processes, Docker Containers, Cloud Foundry Tasks, Kubernetes Jobs, Grid Engine Jobs and more...

_drmaa2 for OS processes and meanwhile more_

[![CircleCI](https://circleci.com/gh/dgruber/drmaa2os.svg?style=svg)](https://circleci.com/gh/dgruber/drmaa2os)

This is a Go API based on an open standard (Open Grid Forum DRMAA2) for submitting and
supervising workloads which can be operating system processes, containers, PODs, tasks,
or batch jobs.

The API allows to develop and run job workflows in OS processes and switch later to 
containers running in Kubernetes, as Cloud Foundry tasks, or pure Docker or
Singularity containers.

Its main pupose is supporting application developers with an abstraction layer on top of 
platforms, workload managers, and cluster schedulers, so that they don't require to deal
with the underlaying details and differences when only simple operations (like starting 
a container and waiting until it is finished) are required. 

It can be easily integrated in applications which create and execute job workflows.

If you are looking for a simple interface for creating job workflows without dealing
with the DRMAA2 details, check out [*wfl*](https://github.com/dgruber/wfl).

For details about the mapping of job operations please consult the platform specific READMEs:

  * [OS Processes](pkg/jobtracker/simpletracker/README.md)
  * [Cloud Foundry](pkg/jobtracker/cftracker/README.md)
  * [Docker / Moby](pkg/jobtracker/dockertracker/README.md)
  * [Kubernetes](pkg/jobtracker/kubernetestracker/README.md)
  * [Singularity](pkg/jobtracker/singularity/README.md)

[Feedback](mailto:info@gridengine.eu) welcome!

For a DRMAA2 implementation based on C DRMAA2 (_libdrmaa2.so_) like for *Univa Grid Engine* please
see [drmaa2](https://github.com/dgruber/drmaa2).

Not yet implemented:

  * [Mesos](pkg/jobtracker/mesostracker/README.md)
  * [C DRMAA Version 1 (libdrmaa.so)](pkg/jobtracker/libdrmaa/README.md)

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





