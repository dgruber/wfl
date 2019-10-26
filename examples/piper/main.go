package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
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

	dir, err := ioutil.TempDir("", "examplepipe")
	if err != nil {
		panic(err)
	}
	defer os.Remove(dir)

	flow.RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "cat",
		Args:          []string{"/etc/services"},
		OutputPath:    filepath.Join(dir, "out"),
	}).ThenRunT(drmaa2interface.JobTemplate{
		RemoteCommand: "sort",
		InputPath:     filepath.Join(dir, "out"),
		OutputPath:    "/dev/stdout",
	}).Wait()
}
