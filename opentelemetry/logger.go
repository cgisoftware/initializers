package opentelemetry

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

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

type LoggerConfig struct {
	context context.Context
	err     error
	httpLog *httpLog
}

type LoggerConfigOption func(d *LoggerConfig)

func WithHttpLog(value *httpLog) LoggerConfigOption {
	return func(c *LoggerConfig) {
		c.httpLog = value
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

	if loggerConfig.httpLog != nil {
		openTelemetryConfig.logger.ErrorContext(ctx, message,
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

		return
	}

	openTelemetryConfig.logger.ErrorContext(ctx, message,
		"error", err,
	)
}
