package db

import (
	"github.com/sealbro/go-feed-me/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqliteDatabase(logger *logger.GormLogger, config *DbConfig) (*DB, error) {
	open, err := gorm.Open(sqlite.Open(config.SqliteConnection), &gorm.Config{
		Logger: logger,
	})

	if err != nil {
		return nil, err
	}

	return &DB{
		DB: open,
	}, err
}
