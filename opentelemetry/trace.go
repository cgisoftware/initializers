package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func StartTracing(ctx context.Context, name string) (context.Context, trace.Span) {
	return openTelemetryConfig.tracer.Start(ctx, name)
}
