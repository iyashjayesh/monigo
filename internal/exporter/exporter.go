package exporter

import (
	"context"
	"errors"

	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/internal/registry"
)

type Exporter interface {
	Export(ctx context.Context, metrics []*registry.MetricValue) error
	Name() string
}

type MultiExporter struct {
	exporters []Exporter
}

func NewMultiExporter(exporters ...Exporter) *MultiExporter {
	return &MultiExporter{exporters: exporters}
}

// Export fans out to all exporters, collecting errors without short-circuiting.
func (m *MultiExporter) Export(ctx context.Context, metrics []*registry.MetricValue) error {
	var errs []error
	for _, e := range m.exporters {
		if err := e.Export(ctx, metrics); err != nil {
			logger.Log.Error("exporter failed", "name", e.Name(), "error", err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Name returns a combined name for the multi-exporter.
func (m *MultiExporter) Name() string {
	return "multi"
}
