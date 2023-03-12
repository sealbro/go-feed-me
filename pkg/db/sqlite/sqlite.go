package sqlite

import (
	"github.com/sealbro/go-feed-me/pkg/db"
	"github.com/sealbro/go-feed-me/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqliteConfig struct {
	Connection string `envconfig:"SQLITE_CONNECTION" default:"feed.db"`
}

func NewSqliteDatabase(logger *logger.GormLogger, config *SqliteConfig) (*db.DB, error) {
	open, err := gorm.Open(sqlite.Open(config.Connection), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}

	sqlite := &db.DB{
		DB: open,
	}

	return sqlite, err
}
