package exporters

import (
	"context"
	"sync"
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

	mu       sync.RWMutex
	gauges   map[string]otelmetric.Float64ObservableGauge
	counters map[string]otelmetric.Float64Counter

	// Latest gauge values, read by callbacks registered once per gauge.
	gaugeValues sync.Map // map[string]gaugeSnapshot
}

type gaugeSnapshot struct {
	value float64
	attrs []attribute.KeyValue
}

// OTelConfig holds configuration for the OTel exporter.
type OTelConfig struct {
	Endpoint string
	Headers  map[string]string
	Insecure bool // When true, use insecure gRPC (default true for backward compat)
}

// NewOTelExporter creates and initializes an OTel OTLP metric exporter.
func NewOTelExporter(ctx context.Context, cfg OTelConfig) (*OTelExporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure || len(cfg.Headers) == 0 {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
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
		gauges:   make(map[string]otelmetric.Float64ObservableGauge),
		counters: make(map[string]otelmetric.Float64Counter),
	}, nil
}

// Export sends metrics to the OTel collector.
// Instruments are created once and reused on subsequent calls.
func (o *OTelExporter) Export(ctx context.Context, metrics []*registry.MetricValue) error {
	var firstErr error
	for _, m := range metrics {
		switch m.Type {
		case registry.Gauge:
			if err := o.exportGauge(m); err != nil && firstErr == nil {
				firstErr = err
			}
		case registry.Counter:
			if err := o.exportCounter(ctx, m); err != nil && firstErr == nil {
				firstErr = err
			}
		case registry.Histogram:
			logger.Log.Warn("histogram export not yet implemented", "metric", m.Name)
		}
	}
	return firstErr
}

func (o *OTelExporter) exportGauge(m *registry.MetricValue) error {
	o.mu.RLock()
	_, exists := o.gauges[m.Name]
	o.mu.RUnlock()

	if !exists {
		o.mu.Lock()
		if _, exists = o.gauges[m.Name]; !exists {
			gauge, err := o.meter.Float64ObservableGauge(m.Name)
			if err != nil {
				o.mu.Unlock()
				logger.Log.Error("failed to create OTel gauge", "metric", m.Name, "error", err)
				return err
			}
			name := m.Name
			_, err = o.meter.RegisterCallback(func(_ context.Context, observer otelmetric.Observer) error {
				if snap, ok := o.gaugeValues.Load(name); ok {
					s := snap.(gaugeSnapshot)
					observer.ObserveFloat64(gauge, s.value, otelmetric.WithAttributes(s.attrs...))
				}
				return nil
			}, gauge)
			if err != nil {
				o.mu.Unlock()
				logger.Log.Error("failed to register OTel callback", "metric", m.Name, "error", err)
				return err
			}
			o.gauges[m.Name] = gauge
		}
		o.mu.Unlock()
	}

	o.gaugeValues.Store(m.Name, gaugeSnapshot{
		value: m.Value,
		attrs: labelsToAttributes(m.Labels),
	})
	return nil
}

func (o *OTelExporter) exportCounter(ctx context.Context, m *registry.MetricValue) error {
	o.mu.RLock()
	counter, exists := o.counters[m.Name]
	o.mu.RUnlock()

	if !exists {
		o.mu.Lock()
		if counter, exists = o.counters[m.Name]; !exists {
			var err error
			counter, err = o.meter.Float64Counter(m.Name)
			if err != nil {
				o.mu.Unlock()
				logger.Log.Error("failed to create OTel counter", "metric", m.Name, "error", err)
				return err
			}
			o.counters[m.Name] = counter
		}
		o.mu.Unlock()
	}

	counter.Add(ctx, m.Value, otelmetric.WithAttributes(labelsToAttributes(m.Labels)...))
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

