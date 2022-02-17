package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/remote/client"
	genclient "github.com/dgruber/drmaa2os/pkg/jobtracker/remote/client/generated"
	"github.com/dgruber/wfl"
)

func main() {

	// with https and basic auth
	ctx := CreateRemoteContextOrPanic()

	flow := wfl.NewWorkflow(ctx).OnError(func(e error) {
		panic("error during workflow creation: " + e.Error())
	})

	flow.Run("sleep", "10").Do(func(j drmaa2interface.Job) {
		fmt.Printf("Started job with ID: %s\n", j.GetID())
	}).OnSuccess(func(j drmaa2interface.Job) {
		fmt.Println("Job finished successfully")
	}).ReapAll()

	fmt.Printf("Submit array job with 10 tasks (1-10) and allow to run max. 2 of them in parallel")
	jobs := flow.RunArrayJob(1, 10, 1, 2, "sleep", "10").Do(func(j drmaa2interface.Job) {
		fmt.Printf("Submitted job array task %s\n", j.GetID())
	})

	fmt.Printf("Waiting for all jobs to finish...\n")
	jobs.Synchronize()
	fmt.Printf("All jobs finished\n")

	if jobs.HasAnyFailed() {
		fmt.Printf("Some jobs failed\n")
		for _, j := range jobs.ListAllFailed() {
			fmt.Printf("Job %s failed\n", j.GetID())
		}
	} else {
		fmt.Printf("All jobs finished successfully\n")
	}
	// remove jobs from DB
	jobs.ReapAll()
}

func CreateRemoteContextOrPanic() *wfl.Context {
	httpsClient, err := getClient("../server/server.crt")
	if err != nil {
		panic(err)
	}
	basicAuthProvider, err := securityprovider.NewSecurityProviderBasicAuth("user", "testpassword")
	if err != nil {
		panic(err)
	}
	params := &client.ClientTrackerParams{
		Server: "https://localhost:8088",
		Path:   "/jobserver/jobmanagement",
		Opts: []genclient.ClientOption{
			genclient.WithHTTPClient(httpsClient),
			genclient.WithRequestEditorFn(basicAuthProvider.Intercept),
		},
	}
	ctx := wfl.NewRemoteContext(wfl.RemoteConfig{}, params)
	if err := ctx.Error(); err != nil {
		panic(err)
	}
	return ctx
}

func getClient(certFile string) (*http.Client, error) {
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cert file: %s", err)
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
		return nil, fmt.Errorf("failed to append cert to system cert pool")
	}
	config := &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		}}, nil
}
