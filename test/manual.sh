#!/bin/sh
CP=/go/src/github.com
HP=/Users/daniel/go/src/github.com

docker run --rm -it -v $HP/dgruber/wfl:$CP/dgruber/wfl -v $HP/dgruber/drmaa2interface:$CP/dgruber/drmaa2interface -v $HP/mitchellh/reflectwalk:$CP/mitchellh/reflectwalk -v $HP/dgruber/drmaa2os:$CP/dgruber/drmaa2os -v $HP/mitchellh/copystructure:$CP/mitchellh/copystructure golang go build -a -v ./src/github.com/dgruber/wfl/examples/docker/docker.go
