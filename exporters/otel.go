package exporters

import (
	"context"
	"time"

	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/internal/registry"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
)

// OTelExporter implements the internal exporter.Exporter interface
// and pushes metrics to an OpenTelemetry Collector via OTLP/gRPC.
type OTelExporter struct {
	provider *metric.MeterProvider
	meter    otelmetric.Meter
	endpoint string
	headers  map[string]string
}

// OTelConfig holds configuration for the OTel exporter.
type OTelConfig struct {
	Endpoint string
	Headers  map[string]string
}

// NewOTelExporter creates and initializes an OTel OTLP metric exporter.
func NewOTelExporter(ctx context.Context, cfg OTelConfig) (*OTelExporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithInsecure(), // TODO: make TLS configurable
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(30*time.Second))),
	)
	meter := provider.Meter("monigo")

	return &OTelExporter{
		provider: provider,
		meter:    meter,
		endpoint: cfg.Endpoint,
		headers:  cfg.Headers,
	}, nil
}

// Export sends metrics to the OTel collector.
func (o *OTelExporter) Export(ctx context.Context, metrics []*registry.MetricValue) error {
	for _, m := range metrics {
		switch m.Type {
		case registry.Gauge:
			gauge, err := o.meter.Float64ObservableGauge(m.Name)
			if err != nil {
				logger.Log.Error("failed to create OTel gauge", "metric", m.Name, "error", err)
				continue
			}
			val := m.Value
			attrs := labelsToAttributes(m.Labels)
			_, err = o.meter.RegisterCallback(func(_ context.Context, observer otelmetric.Observer) error {
				observer.ObserveFloat64(gauge, val, otelmetric.WithAttributes(attrs...))
				return nil
			}, gauge)
			if err != nil {
				logger.Log.Error("failed to register OTel callback", "metric", m.Name, "error", err)
			}

		case registry.Counter:
			counter, err := o.meter.Float64Counter(m.Name)
			if err != nil {
				logger.Log.Error("failed to create OTel counter", "metric", m.Name, "error", err)
				continue
			}
			counter.Add(ctx, m.Value, otelmetric.WithAttributes(labelsToAttributes(m.Labels)...))
		}
	}
	return nil
}

// Name returns the exporter name.
func (o *OTelExporter) Name() string {
	return "otel-otlp"
}

// Shutdown gracefully shuts down the OTel provider.
func (o *OTelExporter) Shutdown(ctx context.Context) error {
	return o.provider.Shutdown(ctx)
}

// labelsToAttributes converts a map of labels to OTel attributes.
func labelsToAttributes(labels map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels))
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
