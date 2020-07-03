package kubernetes

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/wfl"

	// we need to load kubernetes jobtracker
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/kubernetestracker"
)

// Config describes the default container image to use when no other
// is specified in the JobCategory of the JobTemplate. This allows to use the
// Run() method instead of RunT().
type Config struct {
	DefaultImage    string
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewKubernetesContextByCfg creates a new Context with kubernetes as
// task execution engine. The KubernetesConfig configures details, like
// a default container image which is required when Run() is used or
// no JobCategory is set in the JobTemplate.
func NewKubernetesContextByCfg(cfg Config) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	sessionManager, err := drmaa2os.NewKubernetesSessionManager(nil, cfg.DBFile)
	return &wfl.Context{
		SM:                 sessionManager,
		DefaultDockerImage: cfg.DefaultImage,
		CtxCreationErr:     err,
		DefaultTemplate:    cfg.DefaultTemplate,
	}
}

// NewKubernetesContext creates a new Context which executes tasks of
// the workflow in Kubernetes.
func NewKubernetesContext() *wfl.Context {
	return NewKubernetesContextByCfg(Config{})
}
