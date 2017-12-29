# drmaa2os - A Go API for OS Processes, Docker Containers, Cloud Foundry Tasks, Kubernetes Jobs, Grid Engine and more...

This is a Go API based on an open standard (Open Grid Forum DRMAA2) in order to submit and
supervise workloads like OS processes, containers, PODs, tasks from a common interface.

It allows to develop and run job workflows in OS processes, and later easily switch to 
containers running as Cloud Foundry tasks, Docker containers, Grid Engine jobs, etc...

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

For a DRMAA2 implementation based on C DRMAA2 (_libdrmaa2.so_) like for *Univa Grid Engine* please
see [drmaa2](https://github.com/dgruber/drmaa2).

Not yet implemented:

  * [Kubernetes](pkg/jobtracker/kubernetestracker/README.md)
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





