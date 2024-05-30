package db

import (
	"github.com/sealbro/go-feed-me/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
)

func NewPostgresDatabase(logger *logger.GormLogger, config *Config) (*DB, error) {
	open, err := gorm.Open(postgres.Open(config.PostgresConnection), &gorm.Config{
		Logger: logger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: config.PostgresSchema + ".",
		},
	})

	err = open.Use(prometheus.New(prometheus.Config{
		DBName:          config.PostgresSchema,
		RefreshInterval: 15,
		StartServer:     false,
	}))

	if err != nil {
		return nil, err
	}

	db := &DB{
		DB: open,
	}

	return db, err
}
