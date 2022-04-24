package main

import (
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	fmt.Printf("RunMatrix examples\n")
	fmt.Printf("==================\n")

	flow1 := wfl.NewWorkflow(wfl.NewProcessContextByCfg(
		wfl.ProcessConfig{
			DBFile:               "./wfl1.db",
			PersistentJobStorage: false,
		},
	))

	Itereration(flow1)

	flow2 := wfl.NewWorkflow(wfl.NewProcessContextByCfg(
		wfl.ProcessConfig{
			DBFile:               "./wfl2.db",
			PersistentJobStorage: false,
		},
	))

	Matrix(flow2)

}
func Itereration(flow *wfl.Workflow) {
	job := flow.RunMatrixT(drmaa2interface.JobTemplate{
		RemoteCommand: "echo",
		Args:          []string{"{{X}}"},
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stderr",
	}, wfl.Replacement{
		Fields: []wfl.JobTemplateField{
			wfl.Args,
		},
		Pattern:      "{{X}}",
		Replacements: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	}, wfl.Replacement{})
	job.Synchronize()
}

func Matrix(flow *wfl.Workflow) {
	job := flow.RunMatrixT(drmaa2interface.JobTemplate{
		RemoteCommand: "echo",
		Args:          []string{"{{X}} {{Y}}"},
		OutputPath:    "/dev/stdout",
		ErrorPath:     "/dev/stderr",
	}, wfl.Replacement{
		Fields: []wfl.JobTemplateField{
			wfl.Args,
		},
		Pattern:      "{{X}}",
		Replacements: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	}, wfl.Replacement{
		Fields: []wfl.JobTemplateField{
			wfl.Args,
			wfl.JobEnvironment,
		},
		Pattern:      "{{Y}}",
		Replacements: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	})
	job.Synchronize()
}
