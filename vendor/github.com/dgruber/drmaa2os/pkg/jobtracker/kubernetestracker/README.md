# Kubernetes Tracker

Implements the JobTracker interface for kubernetes batch jobs.

## Introduction

The kubernetes tracker provides methods for managing sets of 
grouped batch jobs (JobSessions). JobSessions are implemented
by using labels attached to batch job objects ("drmaa2jobsession")
refering to the JobSession name.

## Functionality

## Notes

At this point in time Kubernetes batch jobs don't play very well with sidecars.
So when using things like _istio_ you might run in state issues (sidecar container
is [still running](https://github.com/istio/istio/issues/6324) after batch job finished).

### Job Control Mapping

| DRMAA2 Job Control | Kubernetes      |
| :-----------------:|:---------------:|
| Suspend            | *Unsupported*   |
| Resume             | *Unsupported*   |
| Terminate          | Delete() - leads to Undetermined state |
| Hold               | *Unsupported*   |
| Release            | *Unsupported*   |

### State Mapping

Based on [JobStatus](https://kubernetes.io/docs/api-reference/batch/v1/definitions/#_v1_jobstatus)

|  DRMAA2 State.                | Kubernetes Job State  |
| :----------------------------:|:---------------------:|
| Done                          | status.Succeeded >= 1 |
| Failed                        | status.Failed >= 1    |
| Suspended                     | -                     |
| Running                       | status.Active >= 1    |
| Queued                        | -                     |
| Undetermined                  | other  / Terminate()  |


### Job Template Mapping

| DRMAA2 JobTemplate   | Kubernetes Batch Job            |
| :-------------------:|:-------------------------------:|
| RemoteCommand        | v1.Container.Command[0]         |
| Args                 | v1.Container.Args               |
| CandidateMachines[0] | v1.Container.Hostname           |
| JobCategory          | v1.Container.Image              |
| WorkingDir           | v1.Container.WorkingDir         |
| JobName              | Note: If set and a job with the same name exists in history submission will fail. metadata: Name |
| DeadlineTime         | AbsoluteTime converted to relative time (v1.Container.ActiveDeadlineSeconds) |

Job Template extensions:

|Extension key  |Extension value                    |
|:--------------|----------------------------------:|
| namespace     | v1.Namespace                      |
| labels        | "key=value,key2=value2" v1.Labels |
 

Required:
* RemoteCommand
* JobCategory as it specifies the image

Other implicit settings:
* Parallelism: 1
* Completions: 1
* BackoffLimit: 1

### Job Info Mapping

| DRMAA2 JobInfo.      | Kubernetes                           |
| :-------------------:|:------------------------------------:|
| ExitStatus           |  0 or 1 (1 if between 1 and 255 / not supported in Status)  |
| SubmissionTime       | job.CreationTimestamp.Time           |
| DispatchTime         | job.Status.StartTime.Time            |
| FinishTime           | job.Status.CompletionTime.Time       |
| State                | see above                            |
| JobID                | v1.Job.UID |

