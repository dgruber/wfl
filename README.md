💙💛

# wfl - A Simple and Pluggable Workflow Language for Go

_Don't mix wfl with [WFL](https://en.wikipedia.org/wiki/Work_Flow_Language)._

[![CircleCI](https://circleci.com/gh/dgruber/wfl/tree/master.svg?style=svg)](https://circleci.com/gh/dgruber/wfl/tree/master)
[![codecov](https://codecov.io/gh/dgruber/wfl/branch/master/graph/badge.svg)](https://codecov.io/gh/dgruber/wfl)

> _Update_: In order to reflect the underlying drmaa2os changes which separates
> different backends more clearly some context creation functions are moved
> to pkg/context. That avoids having to deal with dependencies from bigger libraries
> like Kubernetes or Docker when not using them.

Creating process, container, pod, task, or job workflows based on raw interfaces of
operating systems, Docker, Google Batch, Kubernetes, Cloud Foundry, and HPC job schedulers 
can be a tedios. Lots of repeating code is required. All workload management systems have a
different API.

_wfl_ abstracts away from the underlying details of the processes, containers, and
workload management systems. _wfl_ provides a simple, unified interface which allows
to quickly define and execute a job workflow and change between different execution
backends without changing the workflow itself.

_wfl_ does not come with many features but is simple to use and enough to define and
run jobs and job workflows with inter-job dependencies.

In its simplest form a process can be started and waited for:

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).Run("convert", "image.jpg", "image.png").Wait()
```

If the output of the command needs to be displayed on the terminal you can set the out path in the
default _JobTemplate_ (see below) configuration:

```go
	template := drmaa2interface.JobTemplate{
		ErrorPath:  "/dev/stderr",
		OutputPath: "/dev/stdout",
	}
	flow := wfl.NewWorkflow(wfl.NewProcessContextByCfg(wfl.ProcessConfig{
		DefaultTemplate: template,
	}))
	flow.Run("echo", "hello").Wait()
```

Running a job as a Docker container requires a different context (and the image
already pulled before).

```go
    import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"github.com/dgruber/wfl/pkg/context/docker"
    )
    
    ctx := docker.NewDockerContextByCfg(docker.Config{DefaultDockerImage: "golang:latest"})
    wfl.NewWorkflow(ctx).Run("sleep", "60").Wait()
```

Starting a Docker container without a _run command_ which exposes ports requires more
configuration which can be provided by using a _JobTemplate_ together with the _RunT()_
method.

```go
    jt := drmaa2interface.JobTemplate{
        JobCategory: "swaggerapi/swagger-editor",
    }
    jt.ExtensionList = map[string]string{"exposedPorts": "80:8080/tcp"}
    
    wfl.NewJob(wfl.NewWorkflow(docker.NewDockerContext())).RunT(jt).Wait()
```

Starting a Kubernetes batch job and waiting for its end is not much different.

```go
    wfl.NewWorkflow(kubernetes.NewKubernetesContext()).Run("sleep", "60").Wait()
```

_wfl_ also supports submitting jobs into HPC schedulers like SLURM, Grid Engine and so on.

```go
    wfl.NewWorkflow(libdrmaa.NewLibDRMAAContext()).Run("sleep", "60").Wait()
```

_wfl_ aims to work for any kind of workload. It works on a Mac and Raspberry Pi the same way
as on a high-performance compute cluster. Things missing: On small scale you probably miss data
management - moving results from one job to another. That's deliberately not implemented. But 
some backend implementations (like for Kubernetes) support basic filetransfer in the
_JobTemplate_ (when using _RunT()_) using the _StageInFiles_ and _StageOutFiles_ maps.
On large scale you are missing checkpoint and restart functionality or HA of the workflow 
process itself. Here the idea is not to require any complicated runtime environment
for the workflow applications rather keeping workflows small and repeatably executable
from other workflows.

_wfl_ works with simple primitives: *context*, *workflow*, *job*, and *jobtemplate*

Experimental: Jobs can also be processed in [job control streams](https://github.com/dgruber/wfl/blob/master/examples/stream/stream.go).

First support for logging is also available. Log levels can be controlled by environment variables
(_export WFL_LOGLEVEL=DEBUG_ or _INFO_/_WARNING_/_ERROR_/_NONE_). Applications can use the same
logging facility by getting the logger from the workflow (_workflow.Logger()_) or registering
your own logger in a workflow _(workflow.SetLogger(Logger interface)_). Default is set to ERROR.

### Getting Started

Dependencies of _wfl_ (like drmaa2) are vendored in. The only external package required to be installed
manually is the _drmaa2interface_.

```go
    go get github.com/dgruber/drmaa2interface
```

## Context

A context defines the execution backend for the workflow. Contexts can be easily created
with the _New_ functions which are defined in the _context.go_ file or in the separate
packages found in _pkg/context_.

For creating a context which executes the jobs of a workflow in operating system processses use:

```go
    wfl.NewProcessContext()
```

If the workflow needs to be executed in containers the _DockerContext_ can be used: 

```go
    docker.NewDockerContext()
```

If the Docker context needs to be configured with a default Docker image 
(when Run() is used or RunT() without a configured _JobCategory_ (which _is_ the Docker image))
then the _ContextByCfg()_ can be called.

```go
    docker.NewDockerContextByCfg(docker.Config{DefaultDockerImage: "golang:latest"})
```

For running jobs either in VMs or in containers in Google Batch the _GoogleBatchContext_ needs to be allocated:

```go
    googlebatch.NewGoogleBatchContextByCfg(
		googlebatch.Config{
			DefaultJobCategory: googlebatch.JobCategoryScript, // default container image Run() is using or script if cmd runs as script
			GoogleProjectID:    "google-project",
			Region:             "europe-north1",
			DefaultTemplate: drmaa2interface.JobTemplate{
				MinSlots: 1, // for MPI set MinSlots = MaxSlots and > 1
				MaxSlots: 1, // for just a bunch of tasks MinSlots = 1 (parallelism) and MaxSlots = <tasks>
			},
		})
```  


When you want to run the workflow as Cloud Foundry tasks the _CloudFoundryContext_ can be used:

```go
    cloudfoundry.NewCloudFoundryContext()
```

Without a config it uses following environment variables to access the Cloud Foundry cloud controller API:

* CF_API (like https://api.run.pivotal.io)
* CF_USER
* CF_PASSWORD

For submitting Kubernetes batch jobs a Kubernetes context exists.

```go
   ctx := kubernetes.NewKubernetesContext()
```

Note that each job requires a container image specified which can be done by using
the JobTemplate's JobCategory. When the same container image is used within the whole
job workflow it makes sense to use the Kubernetes config.

```go
   ctx := kubernetes.NewKubernetesContextByCfg(kubernetes.Config{DefaultImage: "busybox:latest"})
```

[Singularity](https://en.wikipedia.org/wiki/Singularity_(software)) containers can be executed
within the Singularity context. When setting the _DefaultImage_ (like in the Kubernetes Context)
then then _Run()_ methods can be used otherwise the Container image must be specified in the 
JobTemplate's _JobCategory_ field separately for each job. The _DefaultImage_
can always be overridden by the _JobCategory_. Note that each task / job
executes a separate Singularity container process.

```go
   ctx := wfl.NewSingularityContextByCfg(wfl.SingularityConfig{DefaultImage: ""}))
```

For working with HPC schedulers the libdrmaa context can be used. This context requires
_libdrmaa.so_ available in the library path at runtime. Grid Engine ships _libdrmaa.so_
but the _LD_LIBRARY_PATH_ needs to be typically set. For SLURM _libdrmaa.so_ often needs
to be [build](https://github.com/natefoo/slurm-drmaa).

Since C go is used under the hood (drmaa2os which uses go drmaa) some compiler flags needs
to be set during build time. Those flags depend on the workload manager used. Best check
out the go drmaa project for finding the right flags.

For building SLURM requires:

    export CGO_LDFLAGS="-L$SLURM_DRMAA_ROOT/lib"
    export CGO_CFLAGS="-DSLURM -I$SLURM_DRMAA_ROOT/include"

If all set a libdrmaa context can be created by importing:

```go
   ctx := libdrmaa.NewLibDRMAAContext()
```

The JobCategory is whatever the workloadmanager associates with it. Typically it is a
set of submission parameters. A basic example is [here](https://github.com/dgruber/wfl/blob/master/examples/libdrmaa/libdrmaa.go).

## Workflow

A workflow encapsulates a set of jobs/tasks using the same backend (context). Depending on the execution
backend it can be seen as a namespace.

It can be created by using:

```go
    wf := wfl.NewWorkflow(ctx)
```

Errors during creation can be catched with

```go
    wf := wfl.NewWorkflow(ctx).OnError(func(e error) {panic(e)})
```

or with

```go
    if wf.HasError() {
        panic(wf.Error())
    }
```

## Job

Jobs are the main objects in _wfl_. A job defines helper methods for dealing with the workload. Many of those methods return the job object itself to allow chaining calls in an easy way. Errors are stored internally and
can be fetched with special methods. A job is as a container and control unit for tasks. Tasks are mapped in most cases to jobs of the underlying workload manager (like in Kubernetes, HPC schedulers etc.) or
raw processes or containers.

The _Run()_ method submits a new task and returns immeadiately, i.e. not waiting for the job to be started
or finished. When the _Run()_ method errors the job submission has failed. The _Wait()_ method waits until the task has been finished. If multiple _Run()_ methods are called in a chain, multiple tasks might be executed
in parallel (depending on the backend). When the same task should be executed multiple times
the _RunArray()_ method might be convinient. When using a HPC workload manager using the
LibDRMAA implementation it gets translated to an array job, which is used for submitting
and running 10s of thousands of tasks in an HPC clusters (like for bioinformatics or for
electronic design automation workloads). Each task gets an unqiue task number set as environment
variable. This is used for accessing specific data sets.

The method _RunMatrixT()_ allows to submit and run multiple tasks based on a job template
with placeholders. Those placeholders get replaced with defined values before jobs get submitted.
That allows to submit many tasks using different job templates in a convinient way
(like for executing a range of commands in a set of different container images for testing).

In some systems it is required to delete job related resources after the job is finished
and no more information needs to be queried about its execution. This functionality is
implemented in the DRMAA2 _Reap()_ method which can be executed by _ReapAll()_ for each
task in the job object. Afterwards the job object should not be used anymore as some
information might not be available anymore. In a Kubernetes environment it removes
the job objects and potentially related objects like configmaps.

Methods can be classified in blocking, non-blocking, job template based, function based, and error handlers.

### Job Submission

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
|  Run() |  Starts a process, container, or submits a task and comes back immediately | no | |
|  RunT() |  Like above but with a JobTemplate as parameter | no | |
|  RunArray() | Submits a bulk job which runs many iterations of the same command | no | |
|  Resubmit() | Submits a job _n_-times (Run().Run().Run()...) | no | |
|  RunEvery() | Submits a task every d _time.Duration_ | yes | |
|  RunEveryT() | Like _RunEvery()_ but with JobTemplate as param | yes | |
|  RunMatrixT() | Replaces placeholders in the job template and submits combinations | no | |

### Job Control

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
| Suspend() | Stops a task from execution (e.g. sending SIGTSTP to the process group)... | | |
| Resume()|  Continues a task (e.g. sending SIGCONT)... | | |
| Kill() | Stops process (SIGKILL), container, task, job immediately. | | |

### Function Execution

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
| Do() | Executes a Go function | yes | |
| Then() | Waits for end of process and executes a Go function | yes | |
| OnSuccess() | Executes a function if the task run successfully (exit code 0)  | yes | |
| OnFailure() | Executes a function if the task failed (exit code != 0)  | yes | |
| OnError() | Executes a function if the task could not be created  | yes | |
| ForAll(f, interface{}) | Executes a user defined function on all tasks | no | |

### Blocker

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
| After() | Blocks a specific amount of time and continues | yes | |
| Wait() | Waits until the task submitted latest finished | yes | |
| Synchronize() | Waits until all submitted tasks finished | yes | |

### Job Flow Control

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
| ThenRun() | Wait() (last task finished) followed by an async Run() | partially | |
| ThenRunT() | ThenRun() with template | partially | |
| OnSuccessRun() | Wait() if Success() then Run() | partially | |
| OnSuccessRunT() | OnSuccessRun() but with template as param | partially | |
| OnFailureRun() | Wait() if Failed() then Run() | partially | |
| OnFailureRunT() | OnFailureRun() but with template as param | partially | |
| Retry() | wait() + !success() + resubmit() + wait() + !success() | yes | |
| AnyFailed() | Cchecks if one of the tasks in the job failed | yes | |

### Job Status and General Checks

| Function Name | Purpose | Blocking | Examples |
| -- | -- | -- | -- |
| JobID() | Returns the ID of the submitted job | no | |
| JobInfo() | Returns the DRMAA2 JobInfo of the job  | no | |
| Template() |   | no | |
| State() |   | no | |
| LastError() |   | no | |
| Failed() |   | no | |
| Success() |   | no | |
| ExitStatus() |   | no | |
| ReapAll() | Cleans up all job related resources from the workload manager. Do not
use the job object afterwards. Calls DRMAA2 Reap() on all tasks. | no | |
| ListAllFailed() | Waits for all tasks and returns the failed tasks as DRMAA2 jobs | yes | |
| ListAll() | Returns all tasks as a slice of DRMAA2 jobs | no | |

## JobTemplate

JobTemplates are specifying the details about a job. In the simplest case the job is specified by the application name and its arguments like it is typically done in the OS shell. In that case the _Run()_ methods (_ThenRun()_, _OnSuccessRun()_, _OnFailureRun()_) can be used. Job template based methods (like _RunT()_) can be completely avoided by providing a
default template when creating the context (_...ByConfig()_). Then each _Run()_ inherits the settings (like _JobCategory_ for the container image name and _OutputPath_ for redirecting output to _stdout_). If more details for specifying the jobs are required the _RunT()_ methods needs to be used.
I'm using currently the [DRMAA2 Go JobTemplate](https://github.com/dgruber/drmaa2interface/blob/master/jobtemplate.go). In most cases only _RemoteCommand_, _Args_, _WorkingDirectory_, _JobCategory_, _JobEnvironment_,  _StageInFiles_ are evaluated. Functionality and semantic is up to the underlying [drmaa2os job tracker](https://github.com/dgruber/drmaa2os/tree/master/pkg/jobtracker).

- [For the process mapping see here](https://github.com/dgruber/drmaa2os/tree/master/pkg/jobtracker/simpletracker)
- [For the mapping to a drmaa1 implementation (libdrmaa.so) for SLURM, Grid Engine, PBS, ...](https://github.com/dgruber/drmaa2os/blob/master/pkg/jobtracker/libdrmaa)
- [For the Docker mapping here](https://github.com/dgruber/drmaa2os/tree/master/pkg/jobtracker/dockertracker)
- [For the Cloud Foundry Task mapping here](https://github.com/dgruber/drmaa2os/blob/master/pkg/jobtracker/cftracker)
- [For the Kubernetes batch job mapping here](https://github.com/dgruber/drmaa2os/blob/master/pkg/jobtracker/kubernetestracker)
- [Singularity support](https://github.com/dgruber/drmaa2os/blob/master/pkg/jobtracker/singularity)

The [_Template_](https://github.com/dgruber/wfl/blob/master/template.go) object provides helper functions for job templates and required as generators of job [streams](https://github.com/dgruber/wfl/blob/master/examples/stream/stream.go). For an example see [here](https://github.com/dgruber/wfl/tree/master/examples/template/template.go).

# Examples

For examples please have a look into the examples directory. [template](https://github.com/dgruber/wfl/tree/master/examples/template/template.go) is a canonical example of a pre-processing job, followed by parallel execution, followed by a post-processing job.

[test](https://github.com/dgruber/wfl/blob/master/test/test.go) is an use case for testing. It compiles
all examples with the local go compiler and then within a Docker container using the _golang:latest_ image
and reports errors.

[cloudfoundry](https://github.com/dgruber/wfl/blob/master/examples/cloudfoundry/cloudfoundry.go) demonstrates how a Cloud Foundry taks can be created.

[Singularity containers](https://github.com/dgruber/wfl/blob/master/examples/singularity/singularity.go) can also be created which is helpful when managing a simple Singularity _wfl_ container workflow within a single HPC job either to fully exploit all resources and reduce the amount of HPC jobs.

## Creating a Workflow which is Executed as OS Processes

The allocated context defines which workload management system / job execution backend is used.

```go
    ctx := wfl.NewProcessContext()
```

Different contexts can be used within a single program. That way multi-clustering potentially
over different cloud solutions is supported.

Using a context a workflow can be established.

```go
    wfl.NewWorkflow(wfl.NewProcessContext())
```

Handling an error during workflow generation can be done by specifying a function which 
is only called in the case of an error.

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).OnError(func(e error) {
		panic(e)
	})
```

The workflow is used in order to instantiate the first job using the _Run()_ method.

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "123")
```

But you can also create an initial job like that:

```go
    job := wfl.NewJob(wfl.NewWorkflow(wfl.NewProcessContext()))
```

For more detailed settings (like resource limits) the DRMAA2 job template can be used as parameter for _RunT()_.

Jobs allow the execution of workload as well as expressing dependencies.

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "2").ThenRun("sleep", "1").Wait()
```

The line above executes two OS processes sequentially and waits until the last job in chain is finished.

In the following example the two sleep processes are executed in parallel. _Wait()_ only waits for the sleep 1 job. Hence sleep 2 still runs after the wait call comes back!

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "2").Run("sleep", "1").Wait()
```

Running two jobs in parallel and waiting until _all jobs_ finished can be done with _Synchronize()_.

```go
    wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "2").Run("sleep", "1").Synchronize()
```

Jobs can also be suspended (stopped) and resumed (continued) - if supported by the execution backend (like OS, Docker).

```go
    wf.Run("sleep", "1").After(time.Millisecond * 100).Suspend().After(time.Millisecond * 100).Resume().Wait()
```

The exit status is available as well. _ExitStatus()_ blocks until the previously submitted job is finished.

```go
    wfl.NewWorkflow(ctx).Run("echo", "hello").ExitStatus()
```

In order to run jobs depending on the exit status the _OnFailure_ and _OnSuccess_ methods can be used:

```go
    wf.Run("false").OnFailureRun("true").OnSuccessRun("false")
```

For executing a function on a submission error _OnError()_ can be used.

For running multiple jobs on a similar job template (like for test workflows) the _RunMatrixT()_
can be used. It expects a _JobTemplate_ with self-defined placeholders (can be any string).
Those placeholders are getting replaced by the lists specified in the Replacements structs.
Then any combination of the replacement lists are evaluated and new job templates are generated
and submitted.

The following example submits and waits for 4 tasks:

* sleep 0.1
* echo 0.1
* sleep 0.2
* echo 0.2

If only a list of replacements is required then the second replacement can just
left empty (_wfl.Replacement{}_). For _JobTemplate_ fields with numbers the replacement
strings are automatically converted to numbers.

```go
job := flow.NewJob().RunMatrixT(
				drmaa2interface.JobTemplate{
					RemoteCommand: "{{cmd}}",
					Args:          []string{"{{arg}}"},
				},
				wfl.Replacement{
					Fields:       []wfl.JobTemplateField{{wfl.RemoteCommand},

					Pattern:      "{{cmd}}",
					Replacements: []string{"sleep", "echo"},
				},
				wfl.Replacement{
					Fields:       []wfl.JobTemplateField{{wfl.Args},

					Pattern:      "{{arg}}",
					Replacements: []string{"0.1", "0.2"},
				},
			)
job.Synchronize()
```

More methods can be found in the sources.

## Basic Workflow Patterns

### Sequence

The successor task runs after the completion of the pre-decessor task.

```go
    flow := wfl.NewWorkflow(ctx)
    flow.Run("echo", "first task").ThenRun("echo", "second task")
    ...
```
or

```go
    flow := wfl.NewWorkflow(ctx)
    job := flow.Run("echo", "first task")
    job.Wait()
    job.Run("echo", "second task")
    ...
```

### Parallel Split

After completion of a task run multiple branches of tasks.

```go

    flow := wfl.NewWorkflow(ctx)
    flow.Run("echo", "first task").Wait()

    notifier := wfl.NewNotifier()

    go func() {
        wfl.NewJob(wfl.NewWorkflow(ctx)).
            TagWith("BranchA").
            Run("sleep", "1").
            ThenRun("sleep", "3").
            Synchronize().
            Notify(notifier)
    }

    go func() {
        wfl.NewJob(wfl.NewWorkflow(ctx)).
            TagWith("BranchB").
            Run("sleep", "1").
            ThenRun("sleep", "3").
            Synchronize().
            Notify(notifier)
    }

    notifier.ReceiveJob()
    notifier.ReceiveJob()

    ...
```

### Synchronization of Tasks

Wait until all tasks of a job which are running in parallel are finished.

```go
    flow := wfl.NewWorkflow(ctx)
    flow.Run("echo", "first task").
        Run("echo", "second task").
        Run("echo", "third task").
        Synchronize()

```

### Synchronization of Branches

Wait until all branches of a workflow are finished.

```go

    notifier := wfl.NewNotifier()

    go func() {
        wfl.NewJob(wfl.NewWorkflow(ctx)).
            TagWith("BranchA").
            Run("sleep", "1").
            Wait().
			Notify(notifier)
    }

    go func() {
        wfl.NewJob(wfl.NewWorkflow(ctx)).
            TagWith("BranchB").
            Run("sleep", "1").
            Wait().
			Notify(notifier)
    }

    notifier.ReceiveJob()
    notifier.ReceiveJob()

    ...
```

### Exclusive Choice

```go
    flow := wfl.NewWorkflow(ctx)
    job := flow.Run("echo", "first task")
    job.Wait()

    if job.Success() {
        // do something
    } else {
        // do something else
    }
    ...
```

### Fork Pattern

When a task is finished _n_ tasks needs to be started in parallel.

```go
    job := wfl.NewWorkflow(ctx).Run("echo", "first task").
        ThenRun("echo", "parallel task 1").
        Run("echo", "parallel task 2").
        Run("echo", "parallel task 3")
    ...
```

or

```go
    flow := wfl.NewWorkflow(ctx)
    
    job := flow.Run("echo", "first task")
    job.Wait()
    for i := 1; i <= 3; i++ {
        job.Run("echo", fmt.Sprintf("parallel task %d", i))
    }
    ...
```

For missing functionality or bugs please open an issue on github. Contributions welcome!
