# Go sigar

## Overview

Go sigar is a golang implementation of the
[sigar API](https://github.com/hyperic/sigar).  The Go version of
sigar has a very similar interface, but is being written from scratch
in pure go/cgo, rather than cgo bindings for libsigar.

## Test drive

    $ go get github.com/cloudfoundry/gosigar
    $ cd $GOPATH/src/github.com/cloudfoundry/gosigar/examples
    $ go run uptime.go

## Supported platforms

Feature | Linux | Darwin | Windows
--------|-------|--------|--------
CPU     |  yes  |   yes  |  yes
Memory  |  yes  |   yes  |  yes
Swap    |  yes  |   yes  |  yes
Load Avg|  yes  |   yes  |  no
FS Usage|  yes  |   yes  |  yes
Disk IO |  yes  |   no   |  yes
Uptime  |  yes  |   yes  |  no
Process |  yes  |   yes  |  no

## License

Apache 2.0
