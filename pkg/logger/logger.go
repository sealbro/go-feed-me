package logger

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type LoggerConfig struct {
}

type Logger struct {
	*otelzap.Logger
}

func NewLogger(_ *LoggerConfig) (*Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	log := otelzap.New(logger, otelzap.WithMinLevel(zap.DebugLevel))

	return &Logger{
		Logger: log,
	}, nil
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
