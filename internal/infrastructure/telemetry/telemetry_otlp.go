package telemetry

import (
	"context"
	"fmt"
	"strings"

	"github.com/SosisterRapStar/flights/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func endpointHostPort(endpoint string) string {
	s := strings.TrimPrefix(endpoint, "http://")
	s = strings.TrimPrefix(s, "https://")
	return s
}

func newOTLPExporter(cfg *config.AppConfig) (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpointHostPort(cfg.Tracing.Endpoint)),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}
	return exp, nil
}

func newResource(cfg *config.AppConfig) *resource.Resource {
	serviceName := cfg.Tracing.ServiceName
	if serviceName == "" {
		serviceName = "flights"
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return resource.Default()
	}
	return res
}
