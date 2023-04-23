package wfl

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"

	// we need to load all the packages for which context creation function
	// are provided so that the code gets registered in the init() functions.
	"github.com/dgruber/drmaa2os/pkg/jobtracker/remote/client"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/simpletracker"

	// need to run Init() to have capabilities available
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/singularity"
)

type SessionManagerType int

// see also drmaa2os package

const (
	// DefaultSessionManager handles jobs as processes
	DefaultSessionManager SessionManagerType = iota
	// DockerSessionManager manages Docker containers
	DockerSessionManager
	// CloudFoundrySessionManager manages Cloud Foundry application tasks
	CloudFoundrySessionManager
	// KubernetesSessionManager creates Kubernetes jobs
	KubernetesSessionManager
	// SingularitySessionManager manages Singularity containers
	SingularitySessionManager
	// SlurmSessionManager manages slurm jobs as cli commands
	SlurmSessionManager
	// LibDRMAASessionManager manages jobs through libdrmaa.so
	LibDRMAASessionManager
	// PodmanSessionManager manages jobs as podman containers either locally or remote
	PodmanSessionManager
	// RemoteSessionManager manages jobs over the network through a remote server
	RemoteSessionManager
	// ExternalSessionManager can be used by external JobTracker implementations
	// during development time before they get added here
	ExternalSessionManager
	// GoogleBatchSessionManager manages Google Cloud Batch jobs
	GoogleBatchSessionManager
	// MPIOperatorSessionManager manages jobs as MPI operator jobs on Kubernetes
	MPIOperatorSessionManager
)

// Context contains a pointer to execution backend and configuration for it.
type Context struct {
	CtxCreationErr     error
	SM                 drmaa2interface.SessionManager
	SMType             SessionManagerType
	DefaultDockerImage string
	// DefaultTemplate contains all default settings for job submission
	// which are copied (if not set) to Run() or RunT() methods
	DefaultTemplate drmaa2interface.JobTemplate
	// ContextTaskID is a number which is incremented for each submitted
	// task. After incrementing and before submitting the task
	// all occurencies of the "{{.ID}}" string in the job template
	// are replaced by the current task ID. Following fields are
	// evaluated: OuputPath, ErrorPath. The workflow can be started
	// with an offset by setting the ContextTaskID to a value > 0.
	ContextTaskID int64
	// Mutext is used for protecting the ContextTaskID
	sync.Mutex
	// JobSessionName is set to "wfl" by default. It can be changed
	// to a custom name. The name is used to create a DRMAA2 session.
	JobSessionName string
}

// WithSessionName set the JobSessionName in the context.
// The name is used to create a DRMAA2 session.
func (c *Context) WithSessionName(jobSessionName string) *Context {
	c.JobSessionName = jobSessionName
	return c
}

// WithUniqueSessionName creates a unique session name which is
// based on the current time and the process ID. Backends with
// persistent job storage (e.g. Docker) would otherwise mix up
// jobs from different application runs in the same flow if the
// same session name is used.
func (c *Context) WithUniqueSessionName() *Context {
	number := rand.NewSource(time.Now().UnixNano()).Int63()
	c.JobSessionName = fmt.Sprintf("wfl-%d-%d", number, os.Getpid())
	return c
}

func (c *Context) WithDefaultDockerImage(image string) *Context {
	c.DefaultDockerImage = image
	return c
}

func (c *Context) WithDefaultJobTemplate(t drmaa2interface.JobTemplate) *Context {
	c.DefaultTemplate = t
	return c
}

func (c *Context) GetNextContextTaskID() int64 {
	c.Lock()
	defer c.Unlock()
	if c.ContextTaskID == math.MaxInt64 {
		c.ContextTaskID = 0
	}
	c.ContextTaskID++
	return c.ContextTaskID
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
	// PersistentJobStorage keeps job state on disk. This slows down
	// job submission but prevents waiting forever for processes which
	// disappeared
	PersistentJobStorage bool
	// JobDBFile is used when PersistentJobStorage is set to true. It must
	// be different from DBFile.
	JobDBFile string
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
	var jobDB string
	if cfg.PersistentJobStorage && cfg.JobDBFile == "" {
		// we need job state DB along with job session DB
		jobDB = TmpFile()
	}
	return NewProcessContextByCfgWithInitParams(ProcessConfig{
		DBFile:          cfg.DBFile,
		DefaultTemplate: cfg.DefaultTemplate},
		simpletracker.SimpleTrackerInitParams{
			UsePersistentJobStorage: cfg.PersistentJobStorage,
			DBFilePath:              jobDB,
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
		SMType:          DefaultSessionManager,
		DefaultTemplate: cfg.DefaultTemplate,
		CtxCreationErr:  err}
}

type RemoteConfig struct {
	LocalDBFile string // job session DB file
	// DefaultTemplate contains the default job submission settings if
	// not overridden by the RunT() like methods.
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewRemoteContext creates a wfl Context for executing jobs through
// a remote connection. The details of the server must be provided in
// the initParams.
func NewRemoteContext(cfg RemoteConfig, initParams *client.ClientTrackerParams) *Context {
	if cfg.LocalDBFile == "" {
		cfg.LocalDBFile = TmpFile()
	}
	sm, err := drmaa2os.NewRemoteSessionManager(*initParams, cfg.LocalDBFile)
	return &Context{
		SM:              sm,
		SMType:          RemoteSessionManager,
		DefaultTemplate: cfg.DefaultTemplate,
		CtxCreationErr:  err}
}

// DRMAA2SessionManagerContext creates a new Context using any given DRMAA2
// Session manager (implementing the drmaa2interface).
func DRMAA2SessionManagerContext(sm drmaa2interface.SessionManager) *Context {
	return &Context{
		SM:             sm,
		SMType:         ExternalSessionManager,
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
// methods are using that container image. That image can be overridden
// by the RunT() method when setting the JobCategory.
func NewSingularityContextByCfg(cfg SingularityConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewSingularitySessionManager(cfg.DBFile)
	return &Context{
		SM:                 sm,
		SMType:             SingularitySessionManager,
		DefaultDockerImage: cfg.DefaultImage,
		CtxCreationErr:     err,
		DefaultTemplate:    cfg.DefaultTemplate,
	}
}
