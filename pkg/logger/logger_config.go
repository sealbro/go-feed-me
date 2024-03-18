package logger

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"INFO"`
}
