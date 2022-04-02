package log

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
)

type Klog struct {
	// logLevelThreshold can be INFO, WARNING, ERROR, NONE
	logLevelThreshold int
}

// NewKlog returns a new instance of a wrapper around klog.
// If printLogLevel is set to WARNING or ERROR the logger
// only prints only logs with the corresponding level or
// higher. If printLogLevel is set to NONE the logger
// prints no logs.
func NewKlogLogger(printLogLevel LogLevel) (Logger, error) {
	var level int
	switch printLogLevel {
	case "INFO":
		level = 1
	case "WARNING":
		level = 2
	case "ERROR":
		level = 3
	case "NONE":
		level = 4
	default:
		return nil, fmt.Errorf("invalid log level: %s", printLogLevel)
	}
	return &Klog{
		logLevelThreshold: level,
	}, nil
}

func (kl *Klog) SetLogLevel(level LogLevel) {
	kl.logLevelThreshold = getLogLevel(level)
	klog.Infof("setting log level to %s", level)
}

func getLogLevel(level LogLevel) int {
	switch level {
	case InfoLevel:
		return 1
	case WarningLevel:
		return 2
	case ErrorLevel:
		return 3
	case NoneLevel:
		return 4
	default:
		return 2
	}
}

// Infof is used for logging at info level.
func (kl *Klog) Infof(ctx context.Context, s string, args ...interface{}) {
	if kl.logLevelThreshold > 1 {
		return
	}
	klog.InfoDepth(getLogDepth(ctx), fmt.Sprintf(s, args...))
}

// getLogDepth returns the log depth of the user function that
// called the log function. Default is 3 but sometimes the
// user function is on a higher level. That needs to be indicated
// by the contexts log-depth value.
func getLogDepth(ctx context.Context) int {
	if ctx == nil {
		return 3
	}
	depth := ctx.Value("log-depth")
	if depth != nil {
		if ds, ok := depth.(int); ok {
			return ds
		}
	}
	return 3
}

// Warningf is used for logging at warning level.
func (kl *Klog) Warningf(ctx context.Context, s string, args ...interface{}) {
	if kl.logLevelThreshold > 2 {
		return
	}
	klog.WarningDepth(getLogDepth(ctx), fmt.Sprintf(s, args...))
}

// Errorf is used for logging at error level.
func (kl *Klog) Errorf(ctx context.Context, s string, args ...interface{}) {
	if kl.logLevelThreshold > 3 {
		return
	}
	klog.ErrorDepth(getLogDepth(ctx), fmt.Sprintf(s, args...))
}

// Begin writes a default log at the begining of a function.
func (kl *Klog) Begin(ctx context.Context, f string) {
	if kl.logLevelThreshold > 1 {
		return
	}
	klog.InfoDepth(getLogDepth(ctx), fmt.Sprintf("Entry: %s", f))
}
