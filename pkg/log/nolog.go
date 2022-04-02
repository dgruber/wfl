package log

import (
	"context"
)

type Nolog struct{}

// NewKlog returns a new instance of a wrapper around klog.
// If printLogLevel is set to WARNING or ERROR the logger
// only prints only logs with the corresponding level or
// higher. If printLogLevel is set to NONE the logger
// prints no logs.
func NewNoLogger() (Logger, error) {
	return &Nolog{}, nil
}

func (l *Nolog) SetLogLevel(ll LogLevel) {
	return
}

// Infof is used for logging at info level.
func (l *Nolog) Infof(ctx context.Context, s string, args ...interface{}) {
	return
}

// Warningf is used for logging at warning level.
func (l *Nolog) Warningf(ctx context.Context, s string, args ...interface{}) {
	return
}

// Errorf is used for logging at error level.
func (l *Nolog) Errorf(ctx context.Context, s string, args ...interface{}) {
	return
}

// Begin writes a default log at the begining of a function.
func (l *Nolog) Begin(ctx context.Context, f string) {
	return
}
