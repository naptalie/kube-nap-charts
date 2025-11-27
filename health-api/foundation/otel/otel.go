// Package otel provides OpenTelemetry configuration for tracing.
package otel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the OpenTelemetry configuration.
type Config struct {
	ServiceName string
	ReporterURI string
	Probability float64
}

// InitTracing initializes OpenTelemetry tracing.
func InitTracing(cfg Config) (trace.Tracer, func(context.Context) error, error) {
	// If no reporter URI is provided, use noop tracer
	if cfg.ReporterURI == "" {
		tracer := trace.NewNoopTracerProvider().Tracer("")
		return tracer, func(context.Context) error { return nil }, nil
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(cfg.ReporterURI),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("creating otlp exporter: %w", err)
	}

	// Set default probability if not specified
	if cfg.Probability == 0 {
		cfg.Probability = 0.05 // 5% sampling
	}

	// Create trace provider
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Probability))),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(100),
			sdktrace.WithBatchTimeout(10*time.Second),
		),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)

	// Set global trace provider
	otel.SetTracerProvider(traceProvider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := traceProvider.Tracer("")

	return tracer, traceProvider.Shutdown, nil
}
