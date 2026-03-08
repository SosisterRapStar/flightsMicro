package telemetry

import (
	"context"
	"sync"

	"github.com/SosisterRapStar/flights/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	provider *sdktrace.TracerProvider
	mu       sync.Mutex
)

func Init(cfg *config.AppConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if cfg == nil || !cfg.Tracing.Enabled {
		return nil
	}

	exporter, err := newOTLPExporter(cfg)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource(cfg)),
	)

	provider = tp
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return nil
}

func Shutdown(ctx context.Context) error {
	mu.Lock()
	tp := provider
	provider = nil
	mu.Unlock()

	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

func Tracer(name string) trace.Tracer {
	mu.Lock()
	tp := provider
	mu.Unlock()

	if tp != nil {
		return tp.Tracer(name)
	}
	return otel.Tracer(name)
}
