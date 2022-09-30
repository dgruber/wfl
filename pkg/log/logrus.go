package log

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// DefaultLogger is the logger used when no other is specified.
type DefaultLogger struct {
	log *logrus.Logger
}

func getDefaultLoggerLevel() logrus.Level {
	switch strings.ToUpper(os.Getenv(logLevelEnv)) {
	case string(DebugLevel):
		return logrus.DebugLevel
	case string(InfoLevel):
		return logrus.InfoLevel
	case string(WarningLevel):
		return logrus.WarnLevel
	case string(ErrorLevel):
		return logrus.ErrorLevel
	case string(NoneLevel):
		return logrus.PanicLevel
	}
	return logrus.WarnLevel
}

// NewDefaultLogger creates the default logger with settings
// found in the process environment.
func NewDefaultLogger() *DefaultLogger {
	l := logrus.New()
	l.Out = os.Stdout
	l.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	l.SetLevel(getDefaultLoggerLevel())
	return &DefaultLogger{
		log: l,
	}
}

func SetLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		logrus.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logrus.SetLevel(logrus.InfoLevel)
	case WarningLevel:
		logrus.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logrus.SetLevel(logrus.ErrorLevel)
	case NoneLevel:
		logrus.SetLevel(logrus.PanicLevel)
	}
}

// Infof is used for logging at info level.
func (dl *DefaultLogger) Infof(ctx context.Context, s string, args ...interface{}) {
	dl.log.Infof(s, args...)
}

// Warningf is used for warning at info level.
func (dl *DefaultLogger) Warningf(ctx context.Context, s string, args ...interface{}) {
	dl.log.Warnf(s, args...)
}

// Errorf is used for errorf at info level.
func (dl *DefaultLogger) Errorf(ctx context.Context, s string, args ...interface{}) {
	dl.log.Errorf(s, args...)
}

// Begin writes a default log at the beginning of a function.
func (dl *DefaultLogger) Begin(ctx context.Context, f string) {
	if ctx == nil {
		ctx = context.Background()
	}
	dl.Infof(ctx, "Entry: %s", f)
}
