package pipeline

import (
	"context"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/internal/exporter"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/internal/registry"
)

type Pipeline struct {
	registry *registry.Registry
	exporter exporter.Exporter
	interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func NewPipeline(r *registry.Registry, e exporter.Exporter, interval time.Duration) *Pipeline {
	return &Pipeline{
		registry: r,
		exporter: e,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (p *Pipeline) Start(ctx context.Context) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := p.registry.GetAll()
				if len(metrics) > 0 {
					if err := p.exporter.Export(ctx, metrics); err != nil {
						logger.Log.Error("pipeline export failed", "exporter", p.exporter.Name(), "error", err)
					}
				}
			case <-p.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop gracefully stops the pipeline. Safe to call multiple times.
func (p *Pipeline) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
	p.wg.Wait()
}
