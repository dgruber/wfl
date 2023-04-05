package cloudfoundry

import (
	"os"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/wfl"

	// we need to load cloudfoundry jobtracker
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/cftracker"
)

// Config descibes where Cloud Foundry (CC API) is found and can be accessed.
type Config struct {
	APIAddr         string
	User            string
	Password        string
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewCloudFoundryContext creates a new Context which allows creating Cloud Foundry
// tasks when executing a Workflow. It reads the configuration out of environment
// variables (CF_API, CF_USER, CF_PASSWORD).
func NewCloudFoundryContext() *wfl.Context {
	cfg := Config{}
	cfg.APIAddr = os.Getenv("CF_API")
	cfg.User = os.Getenv("CF_USER")
	cfg.Password = os.Getenv("CF_PASSWORD")
	cfg.DBFile = ""
	return NewCloudFoundryContextByCfg(cfg)
}

// NewCloudFoundryContextByCfg creates a new task execution Context based on
// the the CloudFoundryContext which describes the API endpoint of the cloud
// controller API of Cloud Foundry.
func NewCloudFoundryContextByCfg(cfg Config) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	sm, err := drmaa2os.NewCloudFoundrySessionManager(cfg.APIAddr, cfg.User, cfg.Password, cfg.DBFile)
	return &wfl.Context{
		SM:              sm,
		SMType:          wfl.CloudFoundrySessionManager,
		CtxCreationErr:  err,
		DefaultTemplate: cfg.DefaultTemplate,
	}
}
