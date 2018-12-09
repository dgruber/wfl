package log

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	logLevelEnv  string = "WFL_LOGLEVEL"
	debugLevel   string = "DEBUG"
	infoLevel    string = "INFO"
	warningLevel string = "WARNING"
	errorLevel   string = "ERROR"
	noneLevel    string = "NONE"
)

// Logger defines all methods required by an logger.
type Logger interface {
	Begin(ctx context.Context, f string)
	Infof(ctx context.Context, s string, args ...interface{})
	Warningf(ctx context.Context, s string, args ...interface{})
	Errorf(ctx context.Context, s string, args ...interface{})
}

// DefaultLogger is the logger used when no other is specified.
type DefaultLogger struct {
	log *logrus.Logger
}

func getDefaultLoggerLevel() logrus.Level {
	switch strings.ToUpper(os.Getenv(logLevelEnv)) {
	case debugLevel:
		return logrus.DebugLevel
	case infoLevel:
		return logrus.InfoLevel
	case warningLevel:
		return logrus.WarnLevel
	case errorLevel:
		return logrus.ErrorLevel
	case noneLevel:
		return logrus.PanicLevel
	}
	return logrus.ErrorLevel
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

// Infof is used for logging at info level.
func (dl *DefaultLogger) Infof(ctx context.Context, s string, args ...interface{}) {
	dl.log.Infof(s, args...)
}

// Warningf is used for logging at info level.
func (dl *DefaultLogger) Warningf(ctx context.Context, s string, args ...interface{}) {
	dl.log.Infof(s, args...)
}

// Errorf is used for logging at info level.
func (dl *DefaultLogger) Errorf(ctx context.Context, s string, args ...interface{}) {
	dl.log.Errorf(s, args...)
}

// Begin writes a default log at the begining of a function.
func (dl *DefaultLogger) Begin(ctx context.Context, f string) {
	dl.Infof(ctx, "Entry: %s", f)
}
