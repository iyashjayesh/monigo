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

func (r *Registry) GetAll() []*MetricValue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	values := make([]*MetricValue, 0, len(r.metrics))
	for _, v := range r.metrics {
		values = append(values, v)
	}
	return values
}
