package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"os"
)

var panicF = func(e error) { panic(e) }

func main() {
	// cf login details needs to be set before in the environment (or create the context ByCfg())
	// CF_API / CF_USER / CF_PASSWORD

	// GUID of app of which the droplet is used as image for the task
	// -> cf app <app_name> --guid
	appGUID := os.Getenv("APP_GUID")

	ctx := wfl.NewCloudFoundryContext().OnError(panicF)

	state := wfl.NewJob(wfl.NewWorkflow(ctx).OnError(panicF)).RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"5"},
		JobName:       "housekeeping",
		JobCategory:   appGUID,
	}).Wait().State()

	fmt.Printf("Cloud Foundry task exited in state %s\n", state)
}
