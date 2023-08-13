package logger

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

type Logger struct {
	*otelzap.Logger
}

func NewLogger(config *LoggerConfig) (*Logger, error) {
	zapConfig := zap.NewProductionConfig()
	level := toZapLevel(config.LogLevel)
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: otelzap.New(logger, otelzap.WithMinLevel(level)),
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

//func createEvent(format string, v []interface{}) *LogEvent {
//	//traceId := trace.SpanFromContext(ctx).SpanContext().TraceID()
//	//if traceId.IsValid() {
//	//	event.TraceId = traceId.String()
//	//}
//
//	return &LogEvent{
//		Source: getSource(2),
//		//Level:     toLogLevelStr(level),
//		Timestamp: time.Now().Format(time.RFC3339),
//		Message:   fmt.Sprintf(format, v...),
//		Hash:      calculateHash(format),
//	}
//}
//
//func getSource(callDepth int) string {
//	_, file, line, ok := runtime.Caller(callDepth + 1)
//	if !ok {
//		file = "???"
//		line = 0
//	}
//
//	split := strings.Split(file, "/")
//	file = split[len(split)-1]
//	return fmt.Sprintf("%s:%d", file, line)
//}
//
//func calculateHash(read string) string {
//	var hashedValue uint64 = 3074457345618258791
//	for _, char := range read {
//		hashedValue += uint64(char)
//		hashedValue *= 3074457345618258799
//	}
//
//	return strings.ToUpper(fmt.Sprintf("%x", hashedValue))
//}
