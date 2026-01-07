package exporters

import (
	"sync"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/prometheus/client_golang/prometheus"
)

// MonigoCollector implements the prometheus.Collector interface.
type MonigoCollector struct {
	cpuUsage    *prometheus.Desc
	memoryUsage *prometheus.Desc
	goroutines  *prometheus.Desc

	diskReadBytes  *prometheus.Desc
	diskWriteBytes *prometheus.Desc

	requestCount *prometheus.Desc
}

var (
	once      sync.Once
	collector *MonigoCollector
)

// NewMonigoCollector returns a singleton instance of MonigoCollector.
func NewMonigoCollector() *MonigoCollector {
	once.Do(func() {
		collector = &MonigoCollector{
			cpuUsage: prometheus.NewDesc(
				"monigo_cpu_usage_percent",
				"Current system CPU usage percentage.",
				nil, nil,
			),
			memoryUsage: prometheus.NewDesc(
				"monigo_memory_usage_bytes",
				"Current system memory usage in bytes.",
				nil, nil,
			),
			goroutines: prometheus.NewDesc(
				"monigo_goroutines_count",
				"Number of goroutines running.",
				nil, nil,
			),
			diskReadBytes: prometheus.NewDesc(
				"monigo_disk_read_bytes_total",
				"Total bytes read from disk.",
				nil, nil,
			),
			diskWriteBytes: prometheus.NewDesc(
				"monigo_disk_write_bytes_total",
				"Total bytes written to disk.",
				nil, nil,
			),
			requestCount: prometheus.NewDesc(
				"monigo_http_requests_total",
				"Total number of HTTP requests processed by MoniGo traced functions.",
				nil, nil,
			),
		}
	})
	return collector
}

// Describe sends the super-set of all possible descriptors of metrics
// collected by this Collector to the provided channel.
func (c *MonigoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsage
	ch <- c.memoryUsage
	ch <- c.goroutines
	ch <- c.diskReadBytes
	ch <- c.diskWriteBytes
	ch <- c.requestCount
}

// Collect is called by the Prometheus registry when collecting metrics.
func (c *MonigoCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch fresh stats from core
	// Note: In a real high-throughput scenario, we might want to cache these
	// or rely on the background strict loop in core. For now, we fetch on scrape.

	// A safer way: Call GetServiceStats which aggregates everything
	stats := core.GetServiceStats()

	// CPU Load
	ch <- prometheus.MustNewConstMetric(
		c.cpuUsage,
		prometheus.GaugeValue,
		common.ParseStringToFloat64(stats.LoadStatistics.SystemCPULoad),
	)

	// Memory Load (bytes)
	// stats.MemoryStatistics.MemoryUsedBySystem is likely "12.5 GB" string.
	// We should probably expose the raw bytes if available, or the load percentage.
	// Let's expose the Load Percentage which is cleaner.
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		common.ParseStringToFloat64(stats.LoadStatistics.SystemMemLoad),
	)

	// Goroutines
	ch <- prometheus.MustNewConstMetric(
		c.goroutines,
		prometheus.GaugeValue,
		float64(stats.CoreStatistics.Goroutines),
	)

	// Disk I/O
	ch <- prometheus.MustNewConstMetric(
		c.diskReadBytes,
		prometheus.CounterValue,
		float64(stats.DiskIO.ReadBytes),
	)
	ch <- prometheus.MustNewConstMetric(
		c.diskWriteBytes,
		prometheus.CounterValue,
		float64(stats.DiskIO.WriteBytes),
	)
}
