#!/bin/bash

cd /opt/sge
./install.sh

source /opt/sge/default/common/settings.sh

export LD_LIBRARY_PATH=$SGE_ROOT/lib/lx-amd64
export PATH=$PATH:/opt/sge/include

# the settings required for libdrmaa Cgo layer
export CGO_LDFLAGS="-L$SGE_ROOT/lib/lx-amd64/"
export CGO_CFLAGS="-DSOG -I$SGE_ROOT/include"

go get -v -d github.com/dgruber/drmaa2interface
go get -v -d github.com/dgruber/wfl

# run tests
cd /go/src/github.com/dgruber/wfl/examples/libdrmaa

go build .

./libdrmaa

exec "$@"
