package db

type Config struct {
	PostgresSchema     string `envconfig:"POSTGRES_SCHEMA" default:"public"`
	PostgresConnection string `envconfig:"POSTGRES_CONNECTION" default:""`
	SqliteConnection   string `envconfig:"SQLITE_CONNECTION" default:"feed.db"`
}
