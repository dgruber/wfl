package libdrmaa

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"

	// we need to load libdrmaa jobtracker (its init method does that for us)
	"github.com/dgruber/drmaa2os/pkg/jobtracker/libdrmaa"
	"github.com/dgruber/wfl"
)

// Config is the configuration for the libdrmaa context.
type Config struct {
	// DBFile is used by the session manager and is created if
	// it does not exist
	DBFile          string
	DefaultTemplate drmaa2interface.JobTemplate
	// JobDBFile when set makes jobs persistent in the DB
	// so that it is possible to re-connect to running jobs
	// after restarting the app (SGE).
	JobDBFile string
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

	if cfg.JobDBFile == "" {
		sm, err := drmaa2os.NewLibDRMAASessionManager(cfg.DBFile)
		return &wfl.Context{
			SM:              sm,
			DefaultTemplate: cfg.DefaultTemplate,
			CtxCreationErr:  err}
	}

	return NewLibDRMAAContextByCfgWithInitParams(cfg, libdrmaa.LibDRMAASessionParams{
		UsePersistentJobStorage: true,
		DBFilePath:              cfg.JobDBFile,
	})
}

// NewLibDRMAAContextByCfg creates a *wfl.Context which is used to manage
// jobs through libdrmaa.so. The configuration accepts a DefaultTemplate
// which can be filled out with default values merged into each job
// submission. Additionally the underlying job tracker can be configured.
// That allows to use features like keeping job ids persistent is a local
// DB.
func NewLibDRMAAContextByCfgWithInitParams(cfg Config, params libdrmaa.LibDRMAASessionParams) *wfl.Context {
	if cfg.DBFile == "" {
		cfg.DBFile = wfl.TmpFile()
	}
	sm, err := drmaa2os.NewLibDRMAASessionManagerWithParams(params, cfg.DBFile)
	return &wfl.Context{
		SM:              sm,
		SMType:          wfl.LibDRMAASessionManager,
		DefaultTemplate: cfg.DefaultTemplate,
		CtxCreationErr:  err}
}
