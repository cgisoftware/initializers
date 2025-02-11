package opentelemetry

import (
	"bytes"
	"context"
	"io"
	"net/http"

	t "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type postgresLog struct {
	query     string
	queryArgs []any
}

type httpLog struct {
	method      string
	path        string
	status      int64
	response    []byte
	headers     map[string][]string
	queryParams map[string][]string
	body        string
	remoteAddr  string
	userAgent   string
}

func NewHttpLog(
	request *http.Request,
	response []byte,
	statusCode int64,
) *httpLog {
	bodyBytes, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return &httpLog{
		method:      request.Method,
		path:        request.URL.Path,
		status:      statusCode,
		response:    response,
		headers:     request.Header,
		queryParams: request.URL.Query(),
		body:        string(bodyBytes),
		remoteAddr:  request.RemoteAddr,
		userAgent:   request.UserAgent(),
	}
}

func NewPostgresLog(query string, queryArgs ...any) *postgresLog {
	return &postgresLog{
		query:     query,
		queryArgs: queryArgs,
	}
}

type LoggerConfig struct {
	context     context.Context
	err         error
	httpLog     *httpLog
	postgresLog *postgresLog
}

type LoggerConfigOption func(d *LoggerConfig)

func WithHttpLog(value *httpLog) LoggerConfigOption {
	return func(c *LoggerConfig) {
		c.httpLog = value
	}
}

func WithPostgresLog(value *postgresLog) LoggerConfigOption {
	return func(c *LoggerConfig) {
		c.postgresLog = value
	}
}

func ErrorLog(
	ctx context.Context,
	message string,
	err error,
	opts ...LoggerConfigOption,
) {
	loggerConfig := &LoggerConfig{
		context: ctx,
		err:     err,
	}
	for _, opt := range opts {
		opt(loggerConfig)
	}

	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	if loggerConfig.httpLog != nil {
		openTelemetryConfig.logger.ErrorContext(ctx, message,
			"traceID", traceID,
			"spanID", spanID,
			"contextName", span.(t.ReadOnlySpan).Name(),
			"error", err,
			"method", loggerConfig.httpLog.method,
			"path", loggerConfig.httpLog.path,
			"status", loggerConfig.httpLog.status,
			"response", string(loggerConfig.httpLog.response),
			"headers", loggerConfig.httpLog.headers,
			"queryParams", loggerConfig.httpLog.queryParams,
			"body", loggerConfig.httpLog.body,
			"remoteAddr", loggerConfig.httpLog.remoteAddr,
			"userAgent", loggerConfig.httpLog.userAgent,
		)
	}

	if loggerConfig.postgresLog != nil {
		openTelemetryConfig.logger.ErrorContext(ctx, message,
			"traceID", traceID,
			"spanID", spanID,
			"contextName", span.(t.ReadOnlySpan).Name(),
			"error", err,
			"query", loggerConfig.postgresLog.query,
			"queryArgs", loggerConfig.postgresLog.queryArgs,
		)
	}
}
