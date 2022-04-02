package log

import (
	"context"
)

type LogLevel string

const (
	logLevelEnv  string   = "WFL_LOGLEVEL"
	DebugLevel   LogLevel = "DEBUG"
	InfoLevel    LogLevel = "INFO"
	WarningLevel LogLevel = "WARNING"
	ErrorLevel   LogLevel = "ERROR"
	NoneLevel    LogLevel = "NONE"
)

// Logger defines all methods required by an logger.
type Logger interface {
	Begin(ctx context.Context, f string)
	Infof(ctx context.Context, s string, args ...interface{})
	Warningf(ctx context.Context, s string, args ...interface{})
	Errorf(ctx context.Context, s string, args ...interface{})
	SetLogLevel(level LogLevel)
}
