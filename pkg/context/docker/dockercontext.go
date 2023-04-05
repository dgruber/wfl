package docker

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/wfl"

	// we need to load kubernetes jobtracker
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/dockertracker"
)

// Config determines configuration options for the Docker containers
// which are created by the Workflow. A common use-case is setting a default
// Docker image, which is used when Run() is called or when RunT() is used
// but the job category is not set in the job template.
type Config struct {
	DBFile             string
	DefaultDockerImage string
	DefaultTemplate    drmaa2interface.JobTemplate
}

// NewDockerContext creates a new Context containing a DRMAA2 session manager
// which is capable for creating Docker containers.
func NewDockerContext() *wfl.Context {
	return NewDockerContextByCfg(Config{DBFile: "", DefaultDockerImage: ""})
}

// NewDockerContextByCfg creates a new Context based on the given DockerConfig.
func NewDockerContextByCfg(cfg Config) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	sm, err := drmaa2os.NewDockerSessionManager(cfg.DBFile)
	return &wfl.Context{
		SM:                 sm,
		SMType:             wfl.DockerSessionManager,
		DefaultDockerImage: cfg.DefaultDockerImage,
		CtxCreationErr:     err,
		DefaultTemplate:    cfg.DefaultTemplate,
	}
}
