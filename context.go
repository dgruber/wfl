package wfl

import (
	"errors"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
)

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

func NewProcessContext() *Context {
	sm, err := drmaa2os.NewDefaultSessionManager("tmp.db")
	return &Context{sm: sm, ctxCreationErr: err}
}

func NewDockerContext(defaultDockerImage, db string) *Context {
	sm, err := drmaa2os.NewDockerSessionManager(db)
	return &Context{sm: sm, defaultDockerImage: defaultDockerImage, ctxCreationErr: err}
}

func NewCloudFoundryContext(addr, username, password, db string) *Context {
	sm, err := drmaa2os.NewCloudFoundrySessionManager(addr, username, password, db)
	return &Context{sm: sm, ctxCreationErr: err}
}

func DRMAA2SessionManagerContext(sm drmaa2interface.SessionManager) *Context {
	return &Context{sm: sm, ctxCreationErr: nil}
}

func ErrorTestContext() *Context {
	return &Context{sm: nil, ctxCreationErr: errors.New("error")}
}
