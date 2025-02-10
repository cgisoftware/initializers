package opentelemetry

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	m "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	t "go.opentelemetry.io/otel/trace"
)

type OpenTelemetryConfig struct {
	context          context.Context
	name             string
	tracer           t.Tracer
	meter            m.Meter
	logger           *slog.Logger
	otelCollectorUri string
	serviceName      string
}

var openTelemetryConfig *OpenTelemetryConfig

type OpenTelemetryOption func(d *OpenTelemetryConfig)

func WithName(value string) OpenTelemetryOption {
	return func(c *OpenTelemetryConfig) {
		c.name = value
	}
}

func WithServiceName(value string) OpenTelemetryOption {
	return func(c *OpenTelemetryConfig) {
		c.serviceName = value
	}
}

func WithOtelCollectorUri(value string) OpenTelemetryOption {
	return func(c *OpenTelemetryConfig) {
		c.otelCollectorUri = value
	}
}

// Initialize retorna um pool de conexões com o banco de dados
func Initialize(ctx context.Context, opts ...OpenTelemetryOption) (shutdown func(context.Context) error, err error) {
	openTelemetryConfig = &OpenTelemetryConfig{
		context:     ctx,
		serviceName: "blank",
		name:        "blank",
	}
	for _, opt := range opts {
		opt(openTelemetryConfig)
	}

	openTelemetryConfig.tracer = otel.Tracer(openTelemetryConfig.name)
	openTelemetryConfig.meter = otel.Meter(openTelemetryConfig.name)
	openTelemetryConfig.logger = otelslog.NewLogger(openTelemetryConfig.name)

	return setupOTelSDK(openTelemetryConfig.context)
}

func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(openTelemetryConfig.serviceName),
	)

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := newTraceProvider(res)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(res)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	loggerProvider, err := newLoggerProvider(res)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(resource *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(openTelemetryConfig.otelCollectorUri),
	)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource),
	)
	return traceProvider, nil
}

func newMeterProvider(resource *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(openTelemetryConfig.otelCollectorUri),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(3*time.Second))),
		metric.WithResource(resource),
	)
	return meterProvider, nil
}

func newLoggerProvider(resource *resource.Resource) (*log.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(
		context.Background(),
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(openTelemetryConfig.otelCollectorUri),
	)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(resource),
	)
	return loggerProvider, nil
}
