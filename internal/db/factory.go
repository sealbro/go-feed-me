package db

import "github.com/sealbro/go-feed-me/pkg/logger"

func NewDatabase(logger *logger.GormLogger, config *Config) (*DB, error) {
	if config.PostgresConnection != "" {
		return NewPostgresDatabase(logger, config)
	} else {
		return NewSqliteDatabase(logger, config)
	}
}
