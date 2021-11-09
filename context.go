package wfl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"

	// we need to load all the packages for which context creation function
	// are provided so that the code gets registered in the init() functions.
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
	// need to run Init() to have capabilities available
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/singularity"
)

// Context contains a pointer to execution backend and configuration for it.
type Context struct {
	CtxCreationErr     error
	SM                 drmaa2interface.SessionManager
	DefaultDockerImage string
	// defaultTemplate contains all default settings for job submission
	// which are copied (if not set) to Run() or RunT() methods
	DefaultTemplate drmaa2interface.JobTemplate
}

// OnError executes a function when an error occurred during
// context creation with the error as parameter.
func (c *Context) OnError(f func(e error)) *Context {
	if c.CtxCreationErr != nil {
		f(c.CtxCreationErr)
	}
	return c
}

// Error returns the error occurred during context creation.
func (c *Context) Error() error {
	return c.CtxCreationErr
}

// HasError returns true if an error during context creation happened.
func (c *Context) HasError() bool {
	return c.CtxCreationErr != nil
}

// TmpFile returns a path to a tmp file in the tmp dir which does not exist yet.
func TmpFile() string {
	var tmpFile string
	for i := 0; i < 1000; i++ {
		rand := fmt.Sprintf("%d%d%d", time.Now().Nanosecond(), os.Getpid(), i)
		tmpFile = filepath.Join(os.TempDir(), fmt.Sprintf("wfl%s.db", rand))
		if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
			break
		}
	}
	if tmpFile == "" {
		panic("could not create tmp workflow database filename")
	}
	return tmpFile
}

// Note that this file contains only context creation functions which don't
// have additional dependencies. Otherwise they get moved to a pkg/context
// subdirectory.

// ProcessConfig contains the configuration for the process context.
type ProcessConfig struct {
	// DBFile is the local file which contains the internal state DB.
	DBFile string
	// DefaultTemplate contains the default job submission settings if
	// not overridden by the RunT() like methods.
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewProcessContext returns a new *Context which manages processes.
func NewProcessContext() *Context {
	return NewProcessContextByCfg(ProcessConfig{
		DBFile:          "",
		DefaultTemplate: drmaa2interface.JobTemplate{}})
}

// NewProcessContextByCfg returns a new *Context which manages processes
// which is configured by the ProcessConfig.
func NewProcessContextByCfg(cfg ProcessConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	return NewProcessContextByCfgWithInitParams(ProcessConfig{
		DBFile:          cfg.DBFile,
		DefaultTemplate: cfg.DefaultTemplate},
		simpletracker.SimpleTrackerInitParams{
			UsePersistentJobStorage: false,
			DBFilePath:              "",
		})
}

// NewProcessContextByCfgWithInitParams returns a new *Context which manages processes
// which is configured by the ProcessConfig.
func NewProcessContextByCfgWithInitParams(cfg ProcessConfig, initParams simpletracker.SimpleTrackerInitParams) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewDefaultSessionManagerWithParams(initParams, cfg.DBFile)
	return &Context{
		SM:              sm,
		DefaultTemplate: cfg.DefaultTemplate,
		CtxCreationErr:  err}
}

// DRMAA2SessionManagerContext creates a new Context using any given DRMAA2
// Session manager (implementing the drmaa2interface).
func DRMAA2SessionManagerContext(sm drmaa2interface.SessionManager) *Context {
	return &Context{
		SM:             sm,
		CtxCreationErr: nil,
	}
}

// ErrorTestContext always returns an error.
func ErrorTestContext() *Context {
	return &Context{
		SM:             nil,
		CtxCreationErr: errors.New("error"),
	}
}

// SingularityConfig contains the default settings for the Singularity
// containers.
type SingularityConfig struct {
	DefaultImage    string
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewSingularityContext creates a new Context which allows to run the
// jobs in Singularity containers. It only works with JobTemplate based
// run methods (like RunT()) as it requires the JobCategory set to the
// the Singularity container image.
func NewSingularityContext() *Context {
	return NewSingularityContextByCfg(SingularityConfig{})
}

// NewSingularityContextByCfg creates a new Context which allows to run
// the jobs in Singularit containers. If the given SingularityConfig
// has set the DefaultImage to valid Singularity image then the Run()
// methods are using that container image. That image can be overriden
// by the RunT() method when setting the JobCategory.
func NewSingularityContextByCfg(cfg SingularityConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewSingularitySessionManager(cfg.DBFile)
	return &Context{
		SM:                 sm,
		DefaultDockerImage: cfg.DefaultImage,
		CtxCreationErr:     err,
		DefaultTemplate:    cfg.DefaultTemplate,
	}
}
