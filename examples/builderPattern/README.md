Example of implementing the Docker builder pattern with _wfl_.

The Go code of the job is in staging/job1/job.go

This file is compiled in a docker image first and the resulting
binary is stored then in an image with only 7mb, which is then
executed.


