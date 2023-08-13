package traces

type JaegerConfig struct {
	AgentHost              string `envconfig:"OTEL_EXPORTER_JAEGER_AGENT_HOST" default:""`
	ApplicationSlug        string `envconfig:"SLUG" default:"feed"`
	ApplicationEnvironment string `envconfig:"ENV" default:"default"`
}
