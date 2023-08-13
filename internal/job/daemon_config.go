package job

type DaemonConfig struct {
	Cron string `envconfig:"CRON" default:"1/60 * * * * *"`
}
