package exporter

import (
	"context"

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

func (m *MultiExporter) Export(ctx context.Context, metrics []*registry.MetricValue) error {
	for _, e := range m.exporters {
		if err := e.Export(ctx, metrics); err != nil {
			// In a real system, we might want to log this and continue
			return err
		}
	}
	return nil
}
