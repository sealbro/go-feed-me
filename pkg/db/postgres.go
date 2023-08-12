package db

import (
	"github.com/sealbro/go-feed-me/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"
)

func NewPostgresDatabase(logger *logger.GormLogger, config *DbConfig) (*DB, error) {
	schemaName := "public"

	open, err := gorm.Open(postgres.Open(config.PostgresConnection), &gorm.Config{
		Logger: logger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: schemaName + ".",
		},
	})

	err = open.Use(prometheus.New(prometheus.Config{
		DBName:          schemaName,
		RefreshInterval: 15,
		//MetricsCollector: []prometheus.MetricsCollector{
		//	&prometheus.Postgres{
		//		VariableNames: []string{"Threads_running"},
		//	},
		//},
	}))

	if err != nil {
		return nil, err
	}

	db := &DB{
		DB: open,
	}

	return db, err
}
