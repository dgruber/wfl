package libdrmaa

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"

	// we need to load libdrmaa jobtracker (its init method does that for us)
	_ "github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa"
	"github.com/dgruber/wfl"
)

// Config is the configuration for the libdrmaa context.
type Config struct {
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewLibDRMAAContext creates a *wfl.Context which is used to manage
// jobs through libdrmaa.so.
func NewLibDRMAAContext() *wfl.Context {
	return NewLibDRMAAContextByCfg(Config{})
}

// NewLibDRMAAContextByCfg creates a *wfl.Context which is used to manage
// jobs through libdrmaa.so. The configuration accepts a DefaultTemplate
// which can be filled out with default values merged into each job
// submission.
func NewLibDRMAAContextByCfg(cfg Config) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	sm, err := drmaa2os.NewLibDRMAASessionManager(cfg.DBFile)
	return &wfl.Context{
		SM:              sm,
		DefaultTemplate: cfg.DefaultTemplate,
		CtxCreationErr:  err}
}
