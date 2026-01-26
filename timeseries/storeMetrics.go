package timeseries

import (
	"fmt"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/nakabonne/tstorage"
)

// GetDataPoints retrieves data points for a given metric and labels.
func GetDataPoints(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error) {
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
	label := tstorage.Label{Name: "host", Value: "server1"}
	var rows []tstorage.Row
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
func generateCoreStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "goroutines",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.Goroutines)},
			Labels:    []tstorage.Label{label},
		},
	}
}

// generateLoadStatsRows generates rows for load statistics.
func generateLoadStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "overall_load_of_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.OverallLoadOfServiceRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "service_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.ServiceCPULoadRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "service_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.ServiceMemLoadRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemCPULoadRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemMemLoadRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_disk_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.SystemDiskLoadRaw},
			Labels:    []tstorage.Label{label},
		},
	}
}

// generateCPUStatsRows generates rows for CPU statistics.
func generateCPUStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "total_cores",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.TotalCores},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "cores_used_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedByService},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "cores_used_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedBySystem},
			Labels:    []tstorage.Label{label},
		},
	}
}

// generateMemoryStatsRows generates rows for memory statistics.
func generateMemoryStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	rows := []tstorage.Row{
		{
			Metric:    "total_system_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.TotalSystemMemoryRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "memory_used_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.MemoryUsedBySystemRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "memory_used_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.MemoryUsedByServiceRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "available_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.AvailableMemoryRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "gc_pause_duration",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.GCPauseDurationRaw},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "stack_memory_usage",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryStatistics.StackMemoryUsageRaw},
			Labels:    []tstorage.Label{label},
		},
	}

	// Adding raw memory statistics records
	for _, record := range serviceMetrics.MemoryStatistics.RawMemStatsRecords {
		rows = append(rows, tstorage.Row{
			Metric:    record.RecordName,
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: record.RecordValue},
			Labels:    []tstorage.Label{label},
		})
	}

	// Adding additional memory statistics
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "heap_alloc_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.HeapAllocByServiceRaw)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "heap_alloc_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.HeapAllocBySystemRaw)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "total_alloc_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.TotalAllocByServiceRaw)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "total_memory_by_os",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.TotalMemoryByOSRaw)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "total_disk_size",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.LoadStatistics.TotalDiskLoadRaw},
			Labels:    []tstorage.Label{label},
		},
	}...)
	return rows
}

// generateNetworkIORows generates rows for network IO statistics.
func generateNetworkIORows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "bytes_sent",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesSent},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "bytes_received",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesReceived},
			Labels:    []tstorage.Label{label},
		},
	}
}

// generateHealthStatsRows generates rows for service and system health statistics.
func generateHealthStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "service_health_percent",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Health.ServiceHealth.Percent},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_health_percent",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Health.SystemHealth.Percent},
			Labels:    []tstorage.Label{label},
		},
	}
}
