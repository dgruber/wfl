# Docker Tracker

## Introduction

Docker Tracker implements the JobTracker interface used by the Go DRMAA2 implementation
in order to use Docker as a backend for managing jobs as containers from the DRMAA2
interface.

## Functionality

## Basic Usage

A JobTemplate requires:
  * JobCategory -> which maps to an installed Docker image
  * RemoteCommand -> which is the command executed within the given Docker image

### Job Control Mapping

| DRMAA2 Job Control | Docker          |
| :-----------------:|:---------------:|
| Suspend            | Signal: SIGSTOP |
| Resume             | Signal: SIGCONT |
| Terminate          | Signal: SIGKILL |
| Hold               | *Unsupported*   |
| Release            | *Unsupported*   |

### State Mapping

| DRMAA2 State                          | Docker State  |
| :------------------------------------:|:-------------:|
| Failed                                | OOMKilled     |
| Failed or Done depending on exit code | Exited        |
| Failed or Done depending on exit code | Dead          |
| Suspended                             | Paused        |
| Running                               | Running       |
| Queued                                | Restarting    |
| Undetermined                          | other         |

## DeleteJob

*DeleteJob* equals *docker rm* and is removing an installed container. It must be terminated / finished before.

### Job Template Mapping

Mapping between the job template and the Docker container config request:

| DRMAA2 JobTemplate   | Docker Container Config Request |
| :-------------------:|:-------------------------------:|
| RemoteCommand        | Cmd[0]                          |
| Args                 | Cmd[1:]                         |
| CandidateMachines[0] | Hostname                        |
| JobCategory          | Image                           |
| WorkingDir           | WorkingDir                      |
| JobEnvironment (k: v)| Env ("k=v")                     |
| StageInFiles         | -v localPath:containerPath      |
| ErrorPath            | Writes stderr into a local file (not a file in container) |
| OutputPath           | Writes stdout into a local file. |
| Extension: "user"    | User / must exist in container if set |
| Extension: "exposedPorts" | -p / multiple entries are splitted with "," |

If more extensions needed just open an issue.

Note that the image must be available (pulled already)!

### Job Info Mapping

### Job Arrays

Since Array Jobs are not supported by Docker the job array functionality is implemented
by creating _n_ tasks sequentially in a loop. The array job ID contains all IDs of the
created Docker containers.

