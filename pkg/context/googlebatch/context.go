package googlebatch

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/wfl"

	"github.com/dgruber/gcpbatchtracker"
)

const (
	// DefaultJobCategory is the default job category used when
	DefaultJobCategory    = gcpbatchtracker.JobCategoryScript
	JobCategoryScript     = gcpbatchtracker.JobCategoryScript
	JobCategoryScriptPath = gcpbatchtracker.JobCategoryScriptPath
)

// Config describes the default container image to use when no other
// is specified in the JobCategory of the JobTemplate. This allows to
// use the Run() method instead of RunT().
type Config struct {
	DBFile             string
	DefaultJobCategory string
	DefaultTemplate    drmaa2interface.JobTemplate
	// Google Project ID mandatory
	GoogleProjectID string
	// Region at GCP mandatory
	Region string
}

// NewGoogleBatchContextByCfg creates a new Context with Google Batch as
// task execution engine.
func NewGoogleBatchContextByCfg(cfg Config) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	if cfg.DefaultTemplate.MinSlots == 0 {
		cfg.DefaultTemplate.MinSlots = 1
	}
	if len(cfg.DefaultTemplate.CandidateMachines) == 0 {
		// default machine type
		cfg.DefaultTemplate.CandidateMachines = []string{"e2-standard-4"}
	}
	sessionManager, err := drmaa2os.NewGoogleBatchSessionManager(
		gcpbatchtracker.GoogleBatchTrackerParams{
			GoogleProjectID: cfg.GoogleProjectID,
			Region:          cfg.Region,
		}, cfg.DBFile)
	return &wfl.Context{
		SM:                 sessionManager,
		SMType:             wfl.GoogleBatchSessionManager,
		DefaultDockerImage: cfg.DefaultJobCategory,
		CtxCreationErr:     err,
		DefaultTemplate:    cfg.DefaultTemplate,
	}
}

// NewGoogleBatchContext creates a new Context which executes tasks of
// the workflow in Google Batch.
func NewGoogleBatchContext(region, googleProjectID string) *wfl.Context {
	return NewGoogleBatchContextByCfg(Config{
		Region:          region,
		GoogleProjectID: googleProjectID,
	})
}
