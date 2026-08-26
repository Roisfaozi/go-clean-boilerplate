package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Config struct {
	ServiceName  string
	CollectorURL string
	Insecure     bool
	SampleRatio  float64
}

// InitTracer initializes the global tracer provider and returns a shutdown function.
func InitTracer(cfg Config) (func(context.Context) error, error) {
	ctx := context.Background()

	// 1. Create Resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 2. Setup OTLP gRPC exporter options
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.CollectorURL),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// 3. Configure Sampler
	ratio := cfg.SampleRatio
	if ratio <= 0 {
		ratio = 1.0
	} else if ratio > 1.0 {
		ratio = 1.0
	}
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))

	// 4. Create Tracer Provider
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// 5. Set Global Objects
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider.Shutdown, nil
}

// StartSpan creates a new span from context.
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	tracer := otel.Tracer("github.com/Roisfaozi/go-clean-boilerplate")
	ctx, span := tracer.Start(ctx, name)
	return ctx, func() {
		span.End()
	}
}
