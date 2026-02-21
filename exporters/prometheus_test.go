package exporters

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMonigoCollector(t *testing.T) {
	c := NewMonigoCollector()
	if c == nil {
		t.Fatal("expected non-nil collector")
	}

	// Singleton check.
	c2 := NewMonigoCollector()
	if c != c2 {
		t.Error("expected singleton instance")
	}
}

func TestDescribe(t *testing.T) {
	c := NewMonigoCollector()
	ch := make(chan *prometheus.Desc, 10)

	go func() {
		c.Describe(ch)
		close(ch)
	}()

	var count int
	for range ch {
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 descriptors, got %d", count)
	}
}

func TestCollect(t *testing.T) {
	c := NewMonigoCollector()
	ch := make(chan prometheus.Metric, 10)

	go func() {
		c.Collect(ch)
		close(ch)
	}()

	var count int
	for range ch {
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 metrics, got %d", count)
	}
}
