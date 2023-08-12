package db

type DbConfig struct {
	PostgresConnection string `envconfig:"POSTGRES_CONNECTION" default:""`
	SqliteConnection   string `envconfig:"SQLITE_CONNECTION" default:"feed.db"`
}
