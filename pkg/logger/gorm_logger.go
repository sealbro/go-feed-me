package logger

import (
	"context"
	"fmt"
	"gorm.io/gorm/logger"
	"log/slog"
	"time"
)

type GormLogger struct {
	logLevel logger.LogLevel
	logger   *Logger
}

func NewGormLogger(logger *Logger) *GormLogger { return &GormLogger{logger: logger} }

func (l *GormLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	l.logLevel = logLevel
	return l
}

func (l *GormLogger) Info(ctx context.Context, message string, values ...any) {
	l.logger.InfoContext(ctx, message, convertToSlog(values)...)
}
func (l *GormLogger) Warn(ctx context.Context, message string, values ...any) {
	l.logger.WarnContext(ctx, message, convertToSlog(values)...)
}
func (l *GormLogger) Error(ctx context.Context, message string, values ...any) {
	l.logger.ErrorContext(ctx, message, convertToSlog(values)...)
}
func (l *GormLogger) Trace(ctx context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), _ error) {
	sql, rows := fc()
	l.logger.DebugContext(ctx, "Sql query", slog.Int64("rows", rows), slog.String("sql", sql))
}

func convertToSlog(input []interface{}) []any {
	var output []any
	for i := 0; i < len(input); i += 2 {
		key, ok := input[i].(string)
		if !ok {
			output = append(output, slog.String("error", fmt.Sprintf("expected string for key, got: %T", input[i])))
			return output
		}
		switch val := input[i+1].(type) {
		case string:
			output = append(output, slog.String(key, val))
		case int:
			output = append(output, slog.Int64(key, int64(val)))
		case float64:
			output = append(output, slog.Float64(key, val))
		case bool:
			output = append(output, slog.Bool(key, val))
		default:
			output = append(output, slog.String("error", fmt.Sprintf("unsupported type for value: %T", input[i+1])))
			return output
		}
	}
	return output
}
