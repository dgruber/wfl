# OS Process Tacker

## Introduction

OS Process Tracker implements the JobTracker interface used by the Go DRMAA2 implementation
in order to use standard OS processes as a backend for managing jobs as processes from the
DRMAA2 interface.

## Functionality

## Basic Usage

A JobTemplate requires at least:
  * RemoteCommand -> which is path to the executable 

### Job Control Mapping

| DRMAA2 Job Control | OS Process      |
| :-----------------:|:---------------:|
| Suspend            |  SIGTSTP        |
| Resume             |  SIGCONT        |
| Terminate          |  SIGKILL        |
| Hold               | *Unsupported*   |
| Release            | *Unsupported*   |

### State Mapping

| DRMAA2 State   | Process State       |
|:--------------:|:-------------------:|
| Queued         | *Unsupported*       |
| Running        | PID is found        |
| Suspended      |                     |
| Done           |                     |
| Failed         |                     |

### DeleteJob


### Job Template Mapping

A JobTemplate is mapped into the process creation process in the following way:

| DRMAA2 JobTemplate   | OS Process                  |
| :-------------------:|:---------------------------:|
| RemoteCommand        | Executable to start         |
| JobName              |                             |
| Args                 | Arguments of the executable |
| WorkingDir           | Working directory           |
| JobEnvironment       | Environment variables set   |
