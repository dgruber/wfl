package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

func main() {
	// Example how to use a temp file for concatenating
	// stdin and stdout of two processes.
	now := time.Now()
	filePipeExample()
	fmt.Printf("file based pipe took %s\n", time.Now().Sub(now).String())
}

func filePipeExample() {
	flow := wfl.NewWorkflow(wfl.NewProcessContext())

	Pipe(flow,
		drmaa2interface.JobTemplate{
			RemoteCommand: "cat",
			Args:          []string{"/etc/services"},
		},
		drmaa2interface.JobTemplate{
			RemoteCommand: "sort",
			OutputPath:    "/dev/stdout",
		},
	)
}

func Pipe(flow *wfl.Workflow, in, out drmaa2interface.JobTemplate) {
	// create a temp file name which in which the first process writes
	// and the second process reads simultaneously
	tmpFile, err := ioutil.TempFile("", "pipe")
	if err != nil {
		panic(err)
	}
	// we are only interested in the unique file name
	tmpFile.Close()
	// remove the file which is filled by the first process when
	// second process has finished
	defer os.Remove(tmpFile.Name())

	in.OutputPath = tmpFile.Name()
	out.InputPath = tmpFile.Name()

	flow.RunT(in).Do(func(j drmaa2interface.Job) {
		// waiting for output file to appear
		// if we don't do that the second process has nothing to read
		// from and will finish immediately
		WaitForFile(tmpFile.Name())
	}).RunT(out).Synchronize()

	os.Remove(tmpFile.Name())
}

func WaitForFile(file string) {
	for {
		if stat, err := os.Stat(file); err == nil {
			if stat.Size() > 0 {
				break
			}
		}
		time.Sleep(time.Millisecond * 2)
	}
}
