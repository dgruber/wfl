# Job Stream

## Example

```go

    func print(j *wfl.Job) *wfl.Job {
	    fmt.Printf("Processing job %s\n", j.JobID())
	    // you can wait for the job here and submit
	    // another task
	    return j
    })

    template := wfl.NewTemplate(drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/sh",
		Args:          []string{"-c", `echo Executing task $TASK_ID`},
    }).AddIterator("tasks", wfl.NewEnvSequenceIterator("TASK_ID", 1, 1))

    config := jstream.Config{
    	Template: template,
    	Workflow: wfl.NewWorkflow(wfl.NewProcessContext()),
    	BufferSize: 16,
    }
    jstream.NewStream(config, nil).Apply(print).Synchronize().Consume()

```

Creates a stream of jobs based on the given configuration and a method which
defines the abort creteria. If set to _nil_ the stream is infinite.

The configuration contains a _Template_ on which _Next()_ is called for getting
a _JobTemplate_ which is submitted with _RunT()_. The configuration also requires
a _workflow_. Optionally a _BufferSize_ can be specified which defines a limit of
how many jobs can be executed in parallel in each step of the stream. Per default
the buffer limit is 0 (due to struct initialization, not because it is a selected
value) which means a new process based on _template.Next()_ can only be executed
if the consumer of the stream takes a task. 

_Synchronize()_ forwards finished (synchronized) jobs. _Consume()_ is required
to remove all jobs from the internal channel (so that the buffer does not blocks)


