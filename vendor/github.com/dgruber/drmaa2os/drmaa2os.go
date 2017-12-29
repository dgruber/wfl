// Package drmaa2os provides a DRMAA2 implementation based on the OS interface.
package drmaa2os

import (
	_ "code.cloudfoundry.org/lager"
	"github.com/dgruber/drmaa2interface"
	_ "os"
)

type Config struct {
	//l *lager.Logger
}

// Listen starts the DRMAA2 OS interface service.
func (c *Config) Listen() error {
	//c.l = lager.NewLogger("drmaa2os")
	//c.l.RegisterSink(*lager.NewWriterSink(os.Stdout, lager.INFO))
	return nil
}

// Stop stops the DRMAA2 OS interface service.
func (c *Config) Stop() error {
	return nil
}

func (c *Config) NewSessionManager() (drmaa2interface.SessionManager, error) {
	/*
		if c == nil || c.db == nil {
			return nil, errors.New("DB not ready.")
		}
		sm := SessionManager{db: c.db, log: l}
	*/
	return &SessionManager{}, nil
}
