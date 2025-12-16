package timeseries

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// Helper function to remove percentage from a string.
func RemovePercentage(s string) float64 {
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val
}

// Helper function to convert a string to a formatted float.
func StringToFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	val, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", val), 64)
	return val
}

// generateCoreStatsRows generates rows for core statistics.
func generateCoreStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    "goroutines",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.Goroutines)},
			Labels:    []tstorage.Label{label},
		},
		// {
		// 	Metric:    "request_count",
		// 	DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.RequestCount)},
		// 	Labels:    []tstorage.Label{label},
		// },
	}
}

// generateLoadStatsRows generates rows for load statistics.
func generateLoadStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {

	return []tstorage.Row{
		{
			Metric:    "overall_load_of_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.OverallLoadOfService)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "service_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.ServiceCPULoad)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "service_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.ServiceMemLoad)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.SystemCPULoad)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "system_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.SystemMemLoad)},
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

func extractFloat(s string) float64 {
	re := regexp.MustCompile(`\d+(\.\d+)?`)
	match := re.FindString(strings.TrimSpace(s))
	if match == "" {
		return 0.0
	}

	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0.0
	}

	return value
}

// generateMemoryStatsRows generates rows for memory statistics.
func generateMemoryStatsRows(serviceMetrics *models.ServiceStats, label tstorage.Label, timestamp int64) []tstorage.Row {
	rows := []tstorage.Row{
		{
			Metric:    "total_system_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.TotalSystemMemory)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "memory_used_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.MemoryUsedBySystem)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "memory_used_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.MemoryUsedByService)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "available_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.AvailableMemory)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "gc_pause_duration",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.GCPauseDuration)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "stack_memory_usage",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.MemoryStatistics.StackMemoryUsage)},
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
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.HeapAllocByService)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "heap_alloc_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.HeapAllocBySystem)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "total_alloc_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.TotalAllocByService)},
			Labels:    []tstorage.Label{label},
		},
		{
			Metric:    "total_memory_by_os",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: extractFloat(serviceMetrics.TotalMemoryByOS)},
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
