package opentelemetry

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func StartTracing(ctx context.Context, name string) (context.Context, trace.Span) {
	return openTelemetryConfig.tracer.Start(ctx, name)
}

func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		verb := request.Method
		path := request.URL.Path
		otelhttp.NewHandler(next, fmt.Sprintf("%s %s", verb, path)).ServeHTTP(response, request)
	})
}
