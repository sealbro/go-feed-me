package traces

type TracesConfig struct {
	OtlpEndpoint           string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:""`
	ApplicationSlug        string `envconfig:"SLUG" default:"feed"`
	ApplicationEnvironment string `envconfig:"ENV" default:"default"`
}
