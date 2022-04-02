package log

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type Zerolog struct {
	logger zerolog.Logger
}

func NewZerologger() (Logger, error) {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	l = l.Level(zerolog.WarnLevel)
	return &Zerolog{
		logger: l,
	}, nil
}

func (l *Zerolog) SetLogLevel(ll LogLevel) {
	l.logger = l.logger.Level(getZeroLogLevel(ll))
	return
}

func getZeroLogLevel(ll LogLevel) zerolog.Level {
	switch ll {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarningLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case NoneLevel:
		return zerolog.PanicLevel
	}
	return zerolog.WarnLevel
}

// Infof is used for logging at info level.
func (l *Zerolog) Infof(ctx context.Context, s string, args ...interface{}) {
	l.logger.Info().Msgf(s, args...)
}

// Warningf is used for logging at warning level.
func (l *Zerolog) Warningf(ctx context.Context, s string, args ...interface{}) {
	l.logger.Warn().Msgf(s, args...)
}

// Errorf is used for logging at error level.
func (l *Zerolog) Errorf(ctx context.Context, s string, args ...interface{}) {
	l.logger.Error().Msgf(s, args...)
}

// Begin writes a default log at the begining of a function.
func (l *Zerolog) Begin(ctx context.Context, f string) {
	l.logger.Info().Msg(f)
}
