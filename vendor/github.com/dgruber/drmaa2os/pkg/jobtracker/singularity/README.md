# Tracker for Singularity Containers

## Introduction

Singularity tracker implements the JobTracker interface used by the Go DRMAA2 implementation
in order to use Singularity containers as a backend for managing jobs as containers from the
DRMAA2 interface on the same host.

## Functionality

Basically the Singularity tracker wraps the OS process tracker adding the required Singularity
calls for creating the process.

## Basic Usage

A JobTemplate requires at least:
  * RemoteCommand -> which is path to the executable
  * JobCategory -> which is the Singularity image (like vsoch-hello-world-master.simg)

If you want to see any output it makes sense to set OutputPath and ErrorPath to /dev/stdout
in the JobTemplate.