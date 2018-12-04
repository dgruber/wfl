# Tracker for Singularity Containers

This is work in progress...ideas welcome!

## Introduction

Singularity tracker implements the JobTracker interface used by the Go DRMAA2 implementation
in order to use Singularity containers as a backend for managing jobs as containers from the
DRMAA2 interface on the same host.

## Functionality

Basically the Singularity tracker wraps the OS process tracker adding the required Singularity
calls for creating the singularity process.

## Basic Usage

A _JobTemplate_ requires at least:
  * RemoteCommand -> which is path to the executable
  * JobCategory -> which is the Singularity image (like vsoch-hello-world-master.simg)

If you want to see any output it makes sense to set OutputPath and ErrorPath to _/dev/stdout_
in the _JobTemplate_.

_JobTemplate_ extensions can be used to inject Singularity exec arguments like "--pid" (see _command.go_).

# Examples

For an example please check out [singularity.go](../../examples/singularity/singularity.go)
in the _examples_ directory.
