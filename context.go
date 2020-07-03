package wfl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"

	// we need to load all the packages so that they get registered
	// TODO initialization needs to be moved to the context package
	// to reduce dependencies
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/cftracker"
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/dockertracker"
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/kubernetestracker"
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/singularity"
)

// Context contains a pointer to execution backend and configuration for it.
type Context struct {
	ctxCreationErr     error
	sm                 drmaa2interface.SessionManager
	defaultDockerImage string
	// defaultTemplate contains all default settings for job submission
	// which are copied (if not set) to Run() or RunT() methods
	defaultTemplate drmaa2interface.JobTemplate
}

// OnError executes a function when an error occurred during
// context creation with the error as parameter.
func (c *Context) OnError(f func(e error)) *Context {
	if c.ctxCreationErr != nil {
		f(c.ctxCreationErr)
	}
	return c
}

// Error returns the error occurred during context creation.
func (c *Context) Error() error {
	return c.ctxCreationErr
}

// HasError returns true if an error during context creation happened.
func (c *Context) HasError() bool {
	return c.ctxCreationErr != nil
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
	sm, err := drmaa2os.NewDefaultSessionManager(cfg.DBFile)
	return &Context{
		sm:              sm,
		defaultTemplate: cfg.DefaultTemplate,
		ctxCreationErr:  err}
}

// DockerConfig determines configuration options for the Docker containers
// which are created by the Workflow. A common use-case is setting a default
// Docker image, which is used when Run() is called or when RunT() is used
// but the job category is not set in the job template.
type DockerConfig struct {
	DBFile             string
	DefaultDockerImage string
	DefaultTemplate    drmaa2interface.JobTemplate
}

// NewDockerContext creates a new Context containing a DRMAA2 session manager
// which is capable for creating Docker containers.
func NewDockerContext() *Context {
	return NewDockerContextByCfg(DockerConfig{DBFile: "", DefaultDockerImage: ""})
}

// NewDockerContextByCfg creates a new Context based on the given DockerConfig.
func NewDockerContextByCfg(cfg DockerConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewDockerSessionManager(cfg.DBFile)
	return &Context{
		sm:                 sm,
		defaultDockerImage: cfg.DefaultDockerImage,
		ctxCreationErr:     err,
		defaultTemplate:    cfg.DefaultTemplate,
	}
}

// CloudFoundryConfig descibes where Cloud Foundry (CC API) is found and can be accessed.
type CloudFoundryConfig struct {
	APIAddr         string
	User            string
	Password        string
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewCloudFoundryContext creates a new Context which allows creating Cloud Foundry
// tasks when executing a Workflow. It reads the configuration out of environment
// variables (CF_API, CF_USER, CF_PASSWORD).
func NewCloudFoundryContext() *Context {
	cfg := CloudFoundryConfig{}
	cfg.APIAddr = os.Getenv("CF_API")
	cfg.User = os.Getenv("CF_USER")
	cfg.Password = os.Getenv("CF_PASSWORD")
	cfg.DBFile = ""
	return NewCloudFoundryContextByCfg(cfg)
}

// NewCloudFoundryContextByCfg creates a new task execution Context based on
// the the CloudFoundryContext which describes the API endpoint of the cloud
// controller API of Cloud Foundry.
func NewCloudFoundryContextByCfg(cfg CloudFoundryConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewCloudFoundrySessionManager(cfg.APIAddr, cfg.User, cfg.Password, cfg.DBFile)
	return &Context{
		sm:              sm,
		ctxCreationErr:  err,
		defaultTemplate: cfg.DefaultTemplate,
	}
}

// DRMAA2SessionManagerContext creates a new Context using any given DRMAA2
// Session manager (implementing the drmaa2interface).
func DRMAA2SessionManagerContext(sm drmaa2interface.SessionManager) *Context {
	return &Context{
		sm:             sm,
		ctxCreationErr: nil,
	}
}

// ErrorTestContext always returns an error.
func ErrorTestContext() *Context {
	return &Context{
		sm:             nil,
		ctxCreationErr: errors.New("error"),
	}
}

// KubernetesConfig describes the default container image to use when no other
// is specified in the JobCategory of the JobTemplate. This allows to use the
// Run() method instead of RunT().
type KubernetesConfig struct {
	DefaultImage    string
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewKubernetesContextByCfg creates a new Context with kubernetes as
// task execution engine. The KubernetesConfig configures details, like
// a default container image which is required when Run() is used or
// no JobCategory is set in the JobTemplate.
func NewKubernetesContextByCfg(cfg KubernetesConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sessionManager, err := drmaa2os.NewKubernetesSessionManager(nil, cfg.DBFile)
	return &Context{
		sm:                 sessionManager,
		defaultDockerImage: cfg.DefaultImage,
		ctxCreationErr:     err,
		defaultTemplate:    cfg.DefaultTemplate,
	}
}

// NewKubernetesContext creates a new Context which executes tasks of
// the workflow in Kubernetes.
func NewKubernetesContext() *Context {
	return NewKubernetesContextByCfg(KubernetesConfig{})
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
	return &Context{sm: sm,
		defaultDockerImage: cfg.DefaultImage,
		ctxCreationErr:     err,
		defaultTemplate:    cfg.DefaultTemplate,
	}
}
