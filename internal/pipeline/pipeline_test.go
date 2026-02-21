package pipeline

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/internal/registry"
)

type mockExporter struct {
	callCount atomic.Int64
	mu        sync.Mutex
	received  [][]*registry.MetricValue
	err       error
}

func (m *mockExporter) Export(_ context.Context, metrics []*registry.MetricValue) error {
	m.callCount.Add(1)
	m.mu.Lock()
	m.received = append(m.received, metrics)
	m.mu.Unlock()
	return m.err
}

func (m *mockExporter) Name() string { return "mock" }

func TestPipelineStartStop(t *testing.T) {
	r := registry.NewRegistry()
	exp := &mockExporter{}
	p := NewPipeline(r, exp, 10*time.Millisecond)

	ctx := context.Background()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	p.Stop()

	// No metrics were set, so exporter should not have been called.
	if exp.callCount.Load() != 0 {
		t.Errorf("expected 0 export calls with empty registry, got %d", exp.callCount.Load())
	}
}

func TestPipelineExportsMetrics(t *testing.T) {
	r := registry.NewRegistry()
	r.SetGauge("test_metric", 42, nil)

	exp := &mockExporter{}
	p := NewPipeline(r, exp, 10*time.Millisecond)

	ctx := context.Background()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	p.Stop()

	if exp.callCount.Load() == 0 {
		t.Error("expected at least one export call")
	}

	exp.mu.Lock()
	defer exp.mu.Unlock()
	if len(exp.received) == 0 || len(exp.received[0]) == 0 {
		t.Fatal("expected received metrics")
	}
	if exp.received[0][0].Name != "test_metric" {
		t.Errorf("expected metric name 'test_metric', got %q", exp.received[0][0].Name)
	}
}

func TestPipelineContextCancellation(t *testing.T) {
	r := registry.NewRegistry()
	exp := &mockExporter{}
	p := NewPipeline(r, exp, 1*time.Hour) // long interval

	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)
	cancel()

	// Stop should return quickly since context was cancelled.
	done := make(chan struct{})
	go func() {
		p.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return after context cancellation")
	}
}

func TestPipelineDoubleStop(t *testing.T) {
	r := registry.NewRegistry()
	exp := &mockExporter{}
	p := NewPipeline(r, exp, 10*time.Millisecond)

	ctx := context.Background()
	p.Start(ctx)
	p.Stop()
	// Should not panic.
	p.Stop()
}
