package logger

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
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
	l.logger.Sugar().Ctx(ctx).Infof(message, values)
}
func (l *GormLogger) Warn(ctx context.Context, message string, values ...interface{}) {
	l.logger.Sugar().Ctx(ctx).Warnf(message, values)
}
func (l *GormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	l.logger.Sugar().Ctx(ctx).Errorf(message, values)
}
func (l *GormLogger) Trace(ctx context.Context, _ time.Time, fc func() (sql string, rowsAffected int64), _ error) {
	sql, rows := fc()

	l.logger.Ctx(ctx).Debug("Sql query", zap.Int64("rows", rows), zap.String("sql", sql))
}
