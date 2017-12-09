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

func (c *Context) OnError(f func(e error)) *Context {
	if c.ctxCreationErr != nil {
		f(c.ctxCreationErr)
	}
	return c
}

func (c *Context) Error() error {
	return c.ctxCreationErr
}

type ProcessConfig struct {
	DBFile string
}

func TmpFile() string {
	var tmpFile string
	for i := 0; i < 1000; i++ {
		rand := fmt.Sprint("%d%d%d", time.Now().Nanosecond(), os.Getpid(), i)
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

func NewProcessContext() *Context {
	return NewProcessContextByCfg(ProcessConfig{DBFile: ""})
}
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
