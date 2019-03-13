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

Note that at this point in time only _JobSessions_ are implemented.

```go
	sm, err := drmaa2os.NewDefaultSessionManager("testdb.db")
	if err != nil {
		panic(err)
	}

	js, err := sm.CreateJobSession("jobsession", "")
	if err != nil {
		panic(err)
	}

	jt := drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"2"},
	}

	job, err := js.RunJob(jt)
	if err != nil {
		panic(err)
	}

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

Using other backends for workload management and execution only differs in creating
a different _SessionManager_. Different _JobTemplate_ attributes might be neccessary when
switching the implementation. If using a backend which supports container images it
might be required to set the _JobCategory_ to the container image name.

### Docker

If Docker is installed locally it will automatically detect it. For pointing to
a different host environment variables needs to be set before the _SessionManager_
is created.

"Use DOCKER_HOST to set the url to the docker server.
 Use DOCKER_API_VERSION to set the version of the API to reach, leave empty for latest.
 Use DOCKER_CERT_PATH to load the TLS certificates from.
 Use DOCKER_TLS_VERIFY to enable or disable TLS verification, off by default."

```go
	sm, err := drmaa2os.NewDockerSessionManager("testdb.db")
	if err != nil {
		panic(err)
	}

	js, err := sm.CreateJobSession("jobsession", "")
	if err != nil {
		panic(err)
	}

	jt := drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"2"},
		JobCategory:   "busybox",
	}
	job, err := js.RunJob(jt)
	if err != nil {
		panic(err)
	}

	job.WaitTerminated(drmaa2interface.InfiniteTime)

	js.Close()
	sm.DestroyJobSession("jobsession")
```

### Kubernetes

```go
	sm, err := drmaa2os.NewKubernetesSessionManager("testdb.db")
	if err != nil {
		panic(err)
	}

	js, err := sm.CreateJobSession("jobsession", "")
	if err != nil {
		panic(err)
	}

	jt := drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"2"},
		JobCategory:   "busybox",
	}
	job, err := js.RunJob(jt)
	if err != nil {
		panic(err)
	}

	job.WaitTerminated(drmaa2interface.InfiniteTime)

	js.Close()
	sm.DestroyJobSession("jobsession")
```

### Cloud Foundry

The Cloud Foundry _SessionManager_ requires details for connecting to the
Cloud Foundry cloud controller API when being created. The _JobCategory_ needs to
be set to the application GUID which is the source of the container image
of the task.

```go
	sm, err := drmaa2os.NewCloudFoundrySessionManager("api.run.pivotal.io", "user", "password", "test.db")
	if err != nil {
		panic(err)
	}

	js, err := sm.CreateJobSession("jobsession", "")
	if err != nil {
		panic(err)
	}

	jt := drmaa2interface.JobTemplate{
		RemoteCommand: "dbbackup.sh",
		Args:          []string{"location"},
		JobCategory:   "123CFAPPGUID",
	}
	job, err := js.RunJob(jt)
	if err != nil {
		panic(err)
	}

	job.WaitTerminated(drmaa2interface.InfiniteTime)

	js.Close()
	sm.DestroyJobSession("jobsession")
```

### Singularity

The Singularity _SessionManager_ wraps the singularity command which is required to be installed.
The container images can be provided in any form (like pointing to file or shub) but are 
required to be set as _JobCategory_ for each job.

```go
	sm, err := drmaa2os.NewSingularitySessionManager("testdb.db")
	if err != nil {
		panic(err)
	}

	js, err := sm.CreateJobSession("jobsession", "")
	if err != nil {
		panic(err)
	}

	jt := drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"2"},
		JobCategory:   "shub://GodloveD/lolcow",
	}

	job, err := js.RunJob(jt)
	if err != nil {
		panic(err)
	}

	job.WaitTerminated(drmaa2interface.InfiniteTime)

	js.Close()
	sm.DestroyJobSession("jobsession")
```




