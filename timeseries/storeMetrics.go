package timeseries

import (
	"fmt"
	"os"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

// GetHostLabel returns a Label with the actual hostname
func GetHostLabel() Label {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return Label{Name: "host", Value: hostname}
}

// GetDataPoints retrieves data points for a given metric and labels.
func GetDataPoints(metric string, labels []Label, start, end int64) ([]DataPoint, error) {
	sto, err := GetStorageInstance()
	if err != nil {
		return nil, fmt.Errorf("error getting storage instance: %w", err)
	}
	return sto.Select(metric, labels, start, end)
}

// StoreServiceMetrics stores service metrics in the time-series storage.
func StoreServiceMetrics(serviceMetrics *models.ServiceStats) error {
	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	location, err := time.LoadLocation("Local")
	if err != nil {
		return fmt.Errorf("error loading location: %w", err)
	}

	currentTime := time.Now().In(location)
	timestamp := currentTime.Unix()
	label := GetHostLabel()
	var rows []Row
	rows = append(rows, generateCoreStatsRows(serviceMetrics, label, timestamp)...)
	rows = append(rows, generateLoadStatsRows(serviceMetrics, label, timestamp)...)
	rows = append(rows, generateCPUStatsRows(serviceMetrics, label, timestamp)...)
	rows = append(rows, generateMemoryStatsRows(serviceMetrics, label, timestamp)...)
	rows = append(rows, generateNetworkIORows(serviceMetrics, label, timestamp)...)
	rows = append(rows, generateHealthStatsRows(serviceMetrics, label, timestamp)...)

	if err := sto.InsertRows(rows); err != nil {
		return fmt.Errorf("error storing service metrics: %w", err)
	}
	return nil
}

// generateCoreStatsRows generates rows for core statistics.
func generateCoreStatsRows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	return []Row{
		{
			Metric:    "goroutines",
			DataPoint: DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.Goroutines)},
			Labels:    []Label{label},
		},
	}
}

// generateLoadStatsRows generates rows for load statistics.
func generateLoadStatsRows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	return []Row{
		{
			Metric:    "overall_load_of_service",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.OverallLoadOfServiceRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "service_cpu_load",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.ServiceCPULoadRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "service_memory_load",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.ServiceMemLoadRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "system_cpu_load",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemCPULoadRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "system_memory_load",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemMemLoadRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "system_disk_load",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemDiskLoadRaw},
			Labels:    []Label{label},
		},
	}
}

// generateCPUStatsRows generates rows for CPU statistics.
func generateCPUStatsRows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	return []Row{
		{
			Metric:    "total_cores",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.TotalCores},
			Labels:    []Label{label},
		},
		{
			Metric:    "cores_used_by_service",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedByService},
			Labels:    []Label{label},
		},
		{
			Metric:    "cores_used_by_system",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedBySystem},
			Labels:    []Label{label},
		},
	}
}

// generateMemoryStatsRows generates rows for memory statistics.
func generateMemoryStatsRows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	rows := []Row{
		{
			Metric:    "total_system_memory",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.TotalSystemMemoryRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "memory_used_by_system",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.MemoryUsedBySystemRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "memory_used_by_service",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.MemoryUsedByServiceRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "available_memory",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.AvailableMemoryRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "gc_pause_duration",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.GCPauseDurationRaw},
			Labels:    []Label{label},
		},
		{
			Metric:    "stack_memory_usage",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.StackMemoryUsageRaw},
			Labels:    []Label{label},
		},
	}

	// Adding raw memory statistics records
	for _, record := range serviceMetrics.MemoryStatistics.RawMemStatsRecords {
		rows = append(rows, Row{
			Metric:    record.RecordName,
			DataPoint: DataPoint{Timestamp: timestamp, Value: record.RecordValue},
			Labels:    []Label{label},
		})
	}

	// Adding additional memory statistics
	rows = append(rows, []Row{
		{
			Metric:    "heap_alloc_by_service",
			DataPoint: DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.HeapAllocByServiceRaw)},
			Labels:    []Label{label},
		},
		{
			Metric:    "heap_alloc_by_system",
			DataPoint: DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.HeapAllocBySystemRaw)},
			Labels:    []Label{label},
		},
		{
			Metric:    "total_alloc_by_service",
			DataPoint: DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.TotalAllocByServiceRaw)},
			Labels:    []Label{label},
		},
		{
			Metric:    "total_memory_by_os",
			DataPoint: DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.TotalMemoryByOSRaw)},
			Labels:    []Label{label},
		},
		{
			Metric:    "total_disk_size",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.TotalDiskLoadRaw},
			Labels:    []Label{label},
		},
	}...)
	return rows
}

// generateNetworkIORows generates rows for network IO statistics.
func generateNetworkIORows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	return []Row{
		{
			Metric:    "bytes_sent",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesSent},
			Labels:    []Label{label},
		},
		{
			Metric:    "bytes_received",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesReceived},
			Labels:    []Label{label},
		},
	}
}

// generateHealthStatsRows generates rows for service and system health statistics.
func generateHealthStatsRows(serviceMetrics *models.ServiceStats, label Label, timestamp int64) []Row {
	return []Row{
		{
			Metric:    "service_health_percent",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.Health.ServiceHealth.Percent},
			Labels:    []Label{label},
		},
		{
			Metric:    "system_health_percent",
			DataPoint: DataPoint{Timestamp: timestamp, Value: serviceMetrics.Health.SystemHealth.Percent},
			Labels:    []Label{label},
		},
	}
}
