package logger

import (
	quartzlogger "github.com/reugn/go-quartz/logger"
	"log"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(config *Config) (*Logger, error) {
	level := toSlogLevel(config.LogLevel)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(OtelHandler{Next: &MsgHashHandler{Next: handler}})

	slog.SetDefault(logger)

	simpleLogger := quartzlogger.NewSimpleLogger(log.Default(), quartzlogger.Level(level))
	quartzlogger.SetDefault(simpleLogger)

	return &Logger{
		Logger: logger,
	}, nil
}

func toSlogLevel(level string) slog.Level {
	switch level {
	case slog.LevelDebug.String():
		return slog.LevelDebug
	case slog.LevelInfo.String():
		return slog.LevelInfo
	case slog.LevelWarn.String():
		return slog.LevelWarn
	case slog.LevelError.String():
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
