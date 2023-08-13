package traces

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	id = 1
)

type ShutdownTracerProvider interface {
	Tracer(name string) trace.Tracer
	Shutdown(ctx context.Context) error
}

type JaegerTracerProvider struct {
	*tracesdk.TracerProvider
}

// NewTraceProvider returns an OpenTelemetry ShutdownTracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// ShutdownTracerProvider will also use a Resource configured with all the information
// about the application.
func NewTraceProvider(config *JaegerConfig) (ShutdownTracerProvider, error) {
	if config.AgentHost == "" {
		return &JaegerTracerProvider{}, nil
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ApplicationSlug),
			attribute.String("environment", config.ApplicationEnvironment),
			attribute.Int64("ID", id),
		)),
	)

	// Register our ShutdownTracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	return &JaegerTracerProvider{tp}, nil
}

func (p *JaegerTracerProvider) Shutdown(ctx context.Context) error {
	if p.TracerProvider == nil {
		return nil
	}

	return p.TracerProvider.Shutdown(ctx)
}

func (p *JaegerTracerProvider) Tracer(name string) trace.Tracer {
	if p.TracerProvider == nil {
		return stubEmptyTracer
	}

	return p.TracerProvider.Tracer(name)
}