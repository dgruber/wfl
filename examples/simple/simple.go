package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

func main() {
	ctx := wfl.NewProcessContext()
	if err := ctx.Error(); err != nil {
		panic(err)
	}
	wfl.NewWorkflow(ctx).OnError(func(e error) {
		panic("error during workflow creation: " + e.Error())
	}).Run("sleep", "0").Do(func(j drmaa2interface.Job) {
		fmt.Printf("Started job with ID: %s\n", j.GetID())
	}).OnSuccess(func(j drmaa2interface.Job) {
		fmt.Println("Job finished successfully")
	})
}
