FROM golang:latest

WORKDIR /go/src/github.com/dgruber/wfl/examples/multi/staging/jobs/JOB

ADD ./staging/jobs/JOB/job.go .

RUN GOOS=linux go build -a -o job .
