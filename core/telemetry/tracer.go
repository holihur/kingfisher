package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func InitTracer(serviceName string) error {
	Tracer = otel.Tracer(serviceName)
	return nil
}

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return Tracer.Start(ctx, name)
}

func Shutdown(ctx context.Context) error { return nil }
