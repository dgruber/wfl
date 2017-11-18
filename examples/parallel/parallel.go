package main

import (
	"github.com/dgruber/wfl"
)

func main() {
	// running 100 "sleep 1s" processes in parallel and wait for all of them
	wfl.NewWorkflow(wfl.NewProcessContext()).Run("sleep", "1").Resubmit(99).Synchronize()
}
