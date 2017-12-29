# Cloud Foundry Tracker

## Introduction

Cloud Foundry Tracker implements the JobTracker interface used by the Go DRMAA2 implementation
in order to use Cloud Foundry tasks as a backend for managing jobs as containers using the
DRMAA2 interface.

## Functionality

## Basic Usage

A JobTemplate requires at least:
  * JobCategory -> which maps to a pushed application GUID
  * RemoteCommand -> which is path to an executable in the container image of the application

### Job Control Mapping

| DRMAA2 Job Control | Cloud Foundry   |
| :-----------------:|:---------------:|
| Suspend            |  *Unsupported*  |
| Resume             |  *Unsupported*  |
| Terminate          |  Terminate Task |
| Hold               |  *Unsupported*  |
| Release            |  *Unsupported*  |

### State Mapping

| DRMAA2 State      |  Cloud Foundry State  |
| :----------------:|:---------------------:|
| Queued            | PENDING               |
| Running           | CANCELING             |
| Running           | RUNNING               |
| Done              | SUCCEEDED             |
| Failed            | FAILED                |

### DeleteJob

Delete job (purging the task information in Cloud Foundry) is not implemented.

### Job Template Mapping

Following mapping between the job template and the Cloud Foundry task request is done:

| DRMAA2 JobTemplate      | Cloud Foundry Task Request |
| :----------------------:|:--------------------------:|
| RemoteCommand           | Command                    |
| JobName                 | Name                       |
| MinPhysMemory (in byte) | MemoryInMegabyte           |
| Args                    | are added to Command       |
| JobCategory             | DropletGUID                |
| WorkingDir              |                            |

