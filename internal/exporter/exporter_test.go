package exporter

import (
	"context"
	"errors"
	"testing"

	"github.com/iyashjayesh/monigo/internal/registry"
)

type fakeExporter struct {
	name string
	err  error
}

func (f *fakeExporter) Export(_ context.Context, _ []*registry.MetricValue) error {
	return f.err
}
func (f *fakeExporter) Name() string { return f.name }

func TestMultiExporterSuccess(t *testing.T) {
	e1 := &fakeExporter{name: "a"}
	e2 := &fakeExporter{name: "b"}
	multi := NewMultiExporter(e1, e2)

	err := multi.Export(context.Background(), []*registry.MetricValue{
		{Name: "test", Value: 1, Type: registry.Gauge},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestMultiExporterContinuesOnError(t *testing.T) {
	e1 := &fakeExporter{name: "a", err: errors.New("e1 failed")}
	e2 := &fakeExporter{name: "b"}
	multi := NewMultiExporter(e1, e2)

	err := multi.Export(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from failing exporter")
	}
	if !errors.Is(err, e1.err) {
		t.Errorf("expected error to contain e1's error, got %v", err)
	}
}

func TestMultiExporterAggregatesErrors(t *testing.T) {
	e1 := &fakeExporter{name: "a", err: errors.New("e1 fail")}
	e2 := &fakeExporter{name: "b", err: errors.New("e2 fail")}
	multi := NewMultiExporter(e1, e2)

	err := multi.Export(context.Background(), nil)
	if err == nil {
		t.Fatal("expected aggregated error")
	}

	errStr := err.Error()
	if !containsSubstring(errStr, "e1 fail") || !containsSubstring(errStr, "e2 fail") {
		t.Errorf("expected both errors in message, got %q", errStr)
	}
}

func TestMultiExporterName(t *testing.T) {
	multi := NewMultiExporter()
	if multi.Name() != "multi" {
		t.Errorf("expected name 'multi', got %q", multi.Name())
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
