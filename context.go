package wfl

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"os"
	"path/filepath"
	"time"
)

// Context contains a pointer to execution backend and configuration for it.
type Context struct {
	ctxCreationErr     error
	sm                 drmaa2interface.SessionManager
	defaultDockerImage string
}

// OnError executes a function when an error occured during
// context creation with the error as parameter.
func (c *Context) OnError(f func(e error)) *Context {
	if c.ctxCreationErr != nil {
		f(c.ctxCreationErr)
	}
	return c
}

// Error returns the error occured during context creation.
func (c *Context) Error() error {
	return c.ctxCreationErr
}

// HasError returns true if an error during context creation happend.
func (c *Context) HasError() bool {
	if c.ctxCreationErr != nil {
		return true
	}
	return false
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
	DBFile string
}

// NewProcessContext returns a new *Context which manages processes.
func NewProcessContext() *Context {
	return NewProcessContextByCfg(ProcessConfig{DBFile: ""})
}

// NewProcessContextByCfg returns a new *Context which manages processes
// which is configured by the ProcessConfig.
func NewProcessContextByCfg(cfg ProcessConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewDefaultSessionManager(cfg.DBFile)
	return &Context{sm: sm, ctxCreationErr: err}
}

type DockerConfig struct {
	DBFile             string
	DefaultDockerImage string
}

func NewDockerContext() *Context {
	return NewDockerContextByCfg(DockerConfig{DBFile: "", DefaultDockerImage: ""})
}

func NewDockerContextByCfg(cfg DockerConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewDockerSessionManager(cfg.DBFile)
	return &Context{sm: sm, defaultDockerImage: cfg.DefaultDockerImage, ctxCreationErr: err}
}

type CloudFoundryConfig struct {
	APIAddr  string
	User     string
	Password string
	DBFile   string
}

func NewCloudFoundryContext() *Context {
	cfg := CloudFoundryConfig{}
	cfg.APIAddr = os.Getenv("CF_API")
	cfg.User = os.Getenv("CF_USER")
	cfg.Password = os.Getenv("CF_PASSWORD")
	cfg.DBFile = ""
	return NewCloudFoundryContextByCfg(cfg)
}

func NewCloudFoundryContextByCfg(cfg CloudFoundryConfig) *Context {
	if cfg.DBFile == "" {
		cfg.DBFile = TmpFile()
	}
	sm, err := drmaa2os.NewCloudFoundrySessionManager(cfg.APIAddr, cfg.User, cfg.Password, cfg.DBFile)
	return &Context{sm: sm, ctxCreationErr: err}
}

func DRMAA2SessionManagerContext(sm drmaa2interface.SessionManager) *Context {
	return &Context{sm: sm, ctxCreationErr: nil}
}

func ErrorTestContext() *Context {
	return &Context{sm: nil, ctxCreationErr: errors.New("error")}
}
