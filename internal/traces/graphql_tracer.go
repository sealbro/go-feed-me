package traces

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/99designs/gqlgen/graphql"
)

type GraphqlTracer struct {
	DisableNonResolverBindingTrace bool
	OperationName                  string
	tracer                         trace.Tracer
}

func NewTraceExtension(provider ShutdownTracerProvider) *GraphqlTracer {
	return &GraphqlTracer{
		tracer: provider.Tracer("graphql"),
	}
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
} = GraphqlTracer{}

func (a GraphqlTracer) ExtensionName() string {
	return "OpenTracing"
}

func (a GraphqlTracer) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

func (a GraphqlTracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)

	operationName := string(oc.Operation.Operation)
	if oc.Operation.Name != "" {
		operationName = fmt.Sprintf("%s_%s", oc.Operation.Operation, oc.Operation.Name)
	}

	tracerCtx, span := a.tracer.Start(ctx, operationName)
	span.AddEvent("log", trace.WithAttributes(attribute.Key("log.query").String(oc.RawQuery)))

	return next(tracerCtx)
}

func (a GraphqlTracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	fc := graphql.GetFieldContext(ctx)

	// Check if this field is disabled
	if a.DisableNonResolverBindingTrace && !fc.IsMethod {
		return next(ctx)
	}

	res, err := next(ctx)

	var attrs []attribute.KeyValue
	if err != nil {
		attrs = append(attrs, attribute.Key("error.message").String(err.Error()))
		attrs = append(attrs, attribute.Key("error.kind").String(fmt.Sprintf("%T", err)))
	}

	errList := graphql.GetFieldErrors(ctx, fc)
	if len(errList) > 0 {
		for idx, err := range errList {
			attrs = append(attrs, attribute.Key(fmt.Sprintf("error.%d.message", idx)).String(err.Error()))
			attrs = append(attrs, attribute.Key(fmt.Sprintf("error.%d.kind", idx)).String(fmt.Sprintf("%T", err)))
		}
	}

	if len(attrs) > 0 {
		span.SetStatus(codes.Error, "GraphQL error")
		span.AddEvent("errors", trace.WithAttributes(attrs...))
	}

	return res, err
}
