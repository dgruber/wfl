#!/bin/bash

docker run --rm -v "$PWD":/usr/src/wfl -w /usr/src/wfl golang:1.16 go build -v && go test -v ./...

