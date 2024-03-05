package logger

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"strings"
)

type LoggerConfig struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

type Logger struct {
	*otelzap.Logger
}

func NewLogger(config *LoggerConfig) (*Logger, error) {
	encodingName := "json_with_hash_encoder"

	err := zap.RegisterEncoder(encodingName, NewHashJSONEncoder)
	if err != nil && !strings.Contains(err.Error(), "encoder already registered") {
		return nil, err
	}

	zapConfig := zap.NewProductionConfig()
	level := toZapLevel(config.LogLevel)
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.Encoding = encodingName
	zapConfig.EncoderConfig.TimeKey = "ts_orig"
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.CallerKey = "source"
	zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	// quartz library uses log.Println
	log.SetOutput(io.Discard)

	tracerLogger := otelzap.New(logger,
		otelzap.WithMinLevel(level),
		otelzap.WithTraceIDField(true))

	return &Logger{
		Logger: tracerLogger,
	}, nil
}

func toZapLevel(level string) zapcore.Level {
	switch level {
	case zap.DebugLevel.String():
		return zap.DebugLevel
	case zap.InfoLevel.String():
		return zap.InfoLevel
	case zap.WarnLevel.String():
		return zap.WarnLevel
	case zap.ErrorLevel.String():
		return zap.ErrorLevel
	case zap.FatalLevel.String():
		return zap.FatalLevel
	case zap.PanicLevel.String():
		return zap.PanicLevel
	default:
		return zap.InfoLevel
	}
}
