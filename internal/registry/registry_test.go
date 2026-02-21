package registry

import (
	"sync"
	"testing"
)

func TestSetGauge(t *testing.T) {
	r := NewRegistry()
	r.SetGauge("cpu", 42.5, map[string]string{"host": "localhost"})

	metrics := r.GetAll()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Name != "cpu" {
		t.Errorf("expected name 'cpu', got %q", metrics[0].Name)
	}
	if metrics[0].Value != 42.5 {
		t.Errorf("expected value 42.5, got %f", metrics[0].Value)
	}
	if metrics[0].Type != Gauge {
		t.Errorf("expected type Gauge, got %d", metrics[0].Type)
	}
	if metrics[0].Labels["host"] != "localhost" {
		t.Errorf("expected label host=localhost, got %q", metrics[0].Labels["host"])
	}
}

func TestSetGaugeOverwrites(t *testing.T) {
	r := NewRegistry()
	r.SetGauge("cpu", 10, nil)
	r.SetGauge("cpu", 20, nil)

	metrics := r.GetAll()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Value != 20 {
		t.Errorf("expected value 20, got %f", metrics[0].Value)
	}
}

func TestIncrementCounter(t *testing.T) {
	r := NewRegistry()
	r.IncrementCounter("requests", 1, map[string]string{"method": "GET"})
	r.IncrementCounter("requests", 5, map[string]string{"method": "GET"})

	metrics := r.GetAll()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Value != 6 {
		t.Errorf("expected value 6, got %f", metrics[0].Value)
	}
	if metrics[0].Type != Counter {
		t.Errorf("expected type Counter, got %d", metrics[0].Type)
	}
}

func TestRecordHistogram(t *testing.T) {
	r := NewRegistry()
	r.RecordHistogram("latency", 0.123, nil)

	metrics := r.GetAll()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Type != Histogram {
		t.Errorf("expected type Histogram, got %d", metrics[0].Type)
	}
}

func TestDelete(t *testing.T) {
	r := NewRegistry()
	r.SetGauge("cpu", 42, nil)
	r.Delete("cpu")

	metrics := r.GetAll()
	if len(metrics) != 0 {
		t.Fatalf("expected 0 metrics after delete, got %d", len(metrics))
	}
}

func TestGetAllReturnsSnapshot(t *testing.T) {
	r := NewRegistry()
	r.SetGauge("cpu", 42, nil)

	metrics := r.GetAll()
	metrics[0].Value = 999

	fresh := r.GetAll()
	if fresh[0].Value != 42 {
		t.Errorf("GetAll should return a copy; original was modified")
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.SetGauge("cpu", float64(i), nil)
			r.IncrementCounter("reqs", 1, nil)
			r.GetAll()
		}(i)
	}
	wg.Wait()

	metrics := r.GetAll()
	if len(metrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(metrics))
	}
}
