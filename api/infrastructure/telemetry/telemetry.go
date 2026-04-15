// Package telemetry menginisialisasi OpenTelemetry SDK dengan:
//   - TracerProvider: OTLP HTTP exporter → Jaeger
//   - MeterProvider: Prometheus exporter → /metrics endpoint
//
// Setelah Init() dipanggil, seluruh pkg yang menggunakan otel.Tracer() dan
// otel.Meter() otomatis menggunakan provider yang sudah dikonfigurasi.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Provider memegang referensi ke TracerProvider dan MeterProvider
// agar bisa di-shutdown dengan graceful saat aplikasi berhenti.
type Provider struct {
	tracer *sdktrace.TracerProvider
	meter  *metric.MeterProvider
}

// Init menginisialisasi OTel SDK dan mendaftarkan global providers.
// otlpEndpoint: URL OTLP collector, e.g. "http://localhost:4318".
// Jika endpoint kosong, trace dikirim ke stdout (no-op exporter diganti stdout).
func Init(ctx context.Context, serviceName, otlpEndpoint string) (*Provider, error) {
	res, err := buildResource(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	tp, err := buildTracerProvider(ctx, res, otlpEndpoint)
	if err != nil {
		return nil, err
	}

	mp, err := buildMeterProvider(res)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	return &Provider{tracer: tp, meter: mp}, nil
}

// Shutdown menutup semua provider secara graceful.
// Harus dipanggil saat aplikasi berhenti (via FX OnStop).
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	if err := p.tracer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("tracer shutdown: %w", err))
	}
	if err := p.meter.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("meter shutdown: %w", err))
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func buildResource(ctx context.Context, serviceName string) (*sdkresource.Resource, error) {
	return sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
		sdkresource.WithOS(),
		sdkresource.WithProcess(),
	)
}

func buildTracerProvider(ctx context.Context, res *sdkresource.Resource, endpoint string) (*sdktrace.TracerProvider, error) {
	opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	if endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, nil
}

func buildMeterProvider(res *sdkresource.Resource) (*metric.MeterProvider, error) {
	exp, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(exp),
		metric.WithResource(res),
	)
	return mp, nil
}
