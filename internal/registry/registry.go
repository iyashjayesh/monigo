package registry

import (
	"sync"
	"time"
)

type MetricType int

const (
	Gauge MetricType = iota
	Counter
	Histogram
)

type MetricValue struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
	Type      MetricType
}

type Registry struct {
	mu      sync.RWMutex
	metrics map[string]*MetricValue
}

func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]*MetricValue),
	}
}

func (r *Registry) SetGauge(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[name] = &MetricValue{
		Name:      name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Type:      Gauge,
	}
}

// IncrementCounter atomically increments a counter metric.
func (r *Registry) IncrementCounter(name string, delta float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.metrics[name]; ok && m.Type == Counter {
		m.Value += delta
		m.Timestamp = time.Now()
	} else {
		r.metrics[name] = &MetricValue{
			Name:      name,
			Value:     delta,
			Labels:    labels,
			Timestamp: time.Now(),
			Type:      Counter,
		}
	}
}

// RecordHistogram records a histogram observation.
// For the OTel exporter this is a no-op placeholder; values are exported as gauges.
func (r *Registry) RecordHistogram(name string, value float64, labels map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[name] = &MetricValue{
		Name:      name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
		Type:      Histogram,
	}
}

// GetAll returns a snapshot copy of all metrics.
func (r *Registry) GetAll() []*MetricValue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	values := make([]*MetricValue, 0, len(r.metrics))
	for _, v := range r.metrics {
		cp := *v
		values = append(values, &cp)
	}
	return values
}

// Delete removes a metric by name.
func (r *Registry) Delete(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
}
