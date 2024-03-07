package logger

import (
	"context"
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

func (l *GormLogger) Info(ctx context.Context, message string, values ...interface{}) {
	l.logger.InfoContext(ctx, message, values)
}
func (l *GormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	l.logger.WarnContext(ctx, message, values)
}
func (l *GormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	l.logger.ErrorContext(ctx, message, values)
}
func (l *GormLogger) Trace(ctx context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), _ error) {
	sql, rows := fc()
	l.logger.InfoContext(ctx, "Sql query", slog.Int64("rows", rows), slog.String("sql", sql))
}
