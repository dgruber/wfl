package main

import (
	"github.com/dgruber/wfl"
)

func main() {
	ctx := wfl.NewProcessContext().OnError(func(err error) {
		panic(err)
	})
	wfl.NewWorkflow(ctx).OnError(func(e error) {
		panic("workflow creation error: " + e.Error())
	}).Run("date", "ivalidformat").OnSuccessRun("touch", "success").OnFailureRun("touch", "failure")
}
