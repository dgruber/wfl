package main

import (
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	"os"
)

var panicF = func(e error) { panic(e) }

func main() {
	// cf login details needs to be set before
	addr := os.Getenv("CF_ADDR")         // like "https://api.run.pivotal.io"
	user := os.Getenv("CF_USER")         // username
	password := os.Getenv("CF_PASSWORD") // password
	// GUID of app of which the droplet is used as image for the task
	// -> cf app <app_name> --guid
	appGUID := os.Getenv("APP_GUID")

	ctx := wfl.NewCloudFoundryContext(addr, user, password, "temp.db").OnError(panicF)
	defer os.Remove("temp.db")

	state := wfl.NewJob(wfl.NewWorkflow(ctx).OnError(panicF)).RunT(drmaa2interface.JobTemplate{
		RemoteCommand: "sleep",
		Args:          []string{"5"},
		JobName:       "housekeeping",
		JobCategory:   appGUID,
	}).Wait().State()

	fmt.Printf("Cloud Foundry task exited in state %s\n", state)
}
