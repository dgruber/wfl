package main

import (
	"fmt"

	"github.com/dgruber/drmaa2interface"
)

func main() {

	testJob := drmaa2interface.JobTemplate{
		JobName:       "testjob",
		RemoteCommand: "./plus.sh",
		InputPath:     "in.txt",
		OutputPath:    "out.txt",
	}

	sortJob := drmaa2interface.JobTemplate{
		JobName:       "sort",
		RemoteCommand: "/usr/bin/sort",
		InputPath:     "/etc/services",
		OutputPath:    "/dev/stdout",
	}

	script := drmaa2interface.JobTemplate{
		RemoteCommand: "/bin/bash",
		Args:          []string{"-c", "echo $JOB_ID && echo $MYVAR"},
		OutputPath:    "/tmp/outputfile.txt",
		JobEnvironment: map[string]string{
			"MYVAR": "myvalue",
		},
	}

	fmt.Print("JSON examples: \n")
	j, _ := testJob.MarshalJSON()
	fmt.Printf("testJob:\n%s\n", string(j))
	j, _ = sortJob.MarshalJSON()
	fmt.Printf("sortJob:\n%s\n", j)
	j, _ = script.MarshalJSON()
	fmt.Printf("script:\n%s\n", j)

}
