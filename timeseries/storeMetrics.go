package timeseries

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/nakabonne/tstorage"
)

// StoreServiceMetrics stores the service metrics in the storage.
func StoreServiceMetrics(serviceMetrics *models.ServiceMetrics) error {
	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	var rows []tstorage.Row
	timestamp := time.Now().Unix()

	rows = []tstorage.Row{
		{
			Metric:    "load_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.Load)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "cores_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.Cores)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "memory_used_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.MemoryUsed)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "number_of_req_served_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.NumberOfReqServerd)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "goroutines_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.GoRoutines)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "total_alloc_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.TotalAlloc)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "memory_alloc_sys_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.MemoryAllocSys)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "heap_alloc_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.HeapAlloc)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "heap_alloc_sys_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.HeapAllocSys)},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "total_duration_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: FormatFloat(serviceMetrics.TotalDurationTookByAPI.Seconds())},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
	}

	if err := sto.InsertRows(rows); err != nil {
		return fmt.Errorf("error storing service metrics: %w", err)
	}

	log.Println("Stored service metrics, timestamp:", timestamp)
	return nil
}

// GetDataPoints retrieves data points for a given metric and labels.
func GetDataPoints(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error) {
	sto, err := GetStorageInstance()
	if err != nil {
		return nil, fmt.Errorf("error getting storage instance: %w", err)
	}

	return sto.Select(metric, labels, start, end)
}

// FormatFloat formats the float value to 2 decimal places.
func FormatFloat(val float64) float64 {
	return float64(int(val*100)) / 100
}

func StoreNewServiceMetrics(serviceMetrics *models.NewServiceStats) error {

	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	var rows []tstorage.Row
	timestamp := time.Now().Unix()

	labelName := "host"
	labelValue := "server1"

	// CoreStatistics
	// GoRoutines & RequestCount
	rows = []tstorage.Row{
		{
			Metric:    "goroutines",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.Goroutines)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "request_count",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.CoreStatistics.RequestCount)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}

	// LoadStatistics
	// overall_load_of_service, service_cpu_load, service_memory_load,
	// system_cpu_load, system_memory_load
	RemovePercentage := func(s string) float64 {
		var val float64
		fmt.Sscanf(s, "%f", &val)
		return val
	}
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "overall_load_of_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.OverallLoadOfService)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "service_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.ServiceCPULoad)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "service_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.ServiceMemLoad)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "system_cpu_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.SystemCPULoad)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "system_memory_load",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: RemovePercentage(serviceMetrics.LoadStatistics.SystemMemLoad)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	// CPUStatistics
	// total_cores, cores_used_by_service, cores_used_by_system
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "total_cores",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.TotalCores},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "cores_used_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedByService},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "cores_used_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.CPUStatistics.CoresUsedBySystem},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	// MemoryStatistics
	// total_system_memory, memory_used_by_system, memory_used_by_service
	// available_memory,
	// gc_pause_duration, stack_memory_usage

	StringToFloat := func(s string) float64 {
		var val float64
		fmt.Sscanf(s, "%f", &val)
		// Formating the value to 4 decimal places
		val, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", val), 64)
		return val
	}

	rows = append(rows, []tstorage.Row{
		{
			Metric:    "total_system_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.TotalSystemMemory)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "memory_used_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.MemoryUsedBySystem)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "memory_used_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.MemoryUsedByService)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "available_memory",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.AvailableMemory)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "gc_pause_duration",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.GCPauseDuration)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "stack_memory_usage",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.MemoryStatistics.StackMemoryUsage)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	// mem_stats_records
	// MemoryStatistics - RawMemStatsRecords
	for _, record := range serviceMetrics.MemoryStatistics.RawMemStatsRecords {
		rows = append(rows, tstorage.Row{
			Metric:    record.RecordName,
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: record.RecordValue},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		})
	}

	// HeapAllocByService
	// HeapAllocBySystem
	// TotalAllocByService
	// TotalMemoryByOS
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "heap_alloc_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.HeapAllocByService)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "heap_alloc_by_system",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.HeapAllocBySystem)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "total_alloc_by_service",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.TotalAllocByService)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "total_memory_by_os",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.TotalMemoryByOS)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	// NetworkIO
	// bytes_sent, bytes_received
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "bytes_sent",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesSent},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
		{
			Metric:    "bytes_received",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.NetworkIO.BytesReceived},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	// OverallHealth
	// overall_health_percent
	rows = append(rows, []tstorage.Row{
		{
			Metric:    "overall_health_percent",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: StringToFloat(serviceMetrics.OverallHealth.OverallHealthPercent)},
			Labels:    []tstorage.Label{{Name: labelName, Value: labelValue}},
		},
	}...)

	if err := sto.InsertRows(rows); err != nil {
		return fmt.Errorf("error storing service metrics: %w", err)
	}

	log.Println("Stored service metrics, timestamp:", timestamp)
	return nil
}
