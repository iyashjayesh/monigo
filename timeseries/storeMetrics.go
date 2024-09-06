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

// {
//     "core_statistics": {
//         "goroutines": 8,
//         "request_count": 0,
//         "uptime": "8.67 s",
//         "total_duration_took_by_request": 0
//     },
//     "load_statistics": {
//         "service_cpu_load": "0.41%",
//         "system_cpu_load": "32.27%",
//         "total_cpu_load": "33.10%",
//         "service_memory_load": "0.13%",
//         "system_memory_load": "73.78%",
//         "total_memory_load": "16.00 GB",
//         "overall_load_of_service": "0.27%"
//     },
//     "cpu_statistics": {
//         "total_cores": 10,
//         "total_logical_cores": 10,
//         "cores_used_by_system": 3.402,
//         "cores_used_by_service": 0.047,
//         "cores_used_by_service_in_percent": "0.05%",
//         "cores_used_by_system_in_percent": "3.40%"
//     },
//     "memory_statistics": {
//         "total_system_memory": "16.00 GB",
//         "memory_used_by_system": "11.46 GB",
//         "memory_used_by_service": "5.08 MB",
//         "available_memory": "4.54 GB",
//         "gc_pause_duration": "0.06 ms",
//         "stack_memory_usage": "544.00 KB",
//         "total_swap_memory": "0.00 B",
//         "free_swap_memory": "0.00 B",
//         "mem_stats_records": [
//             {
//                 "record_name": "Alloc",
//                 "record_description": "Bytes of allocated heap objects.",
//                 "record_value": 5.075103759765625,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "TotalAlloc",
//                 "record_description": "Cumulative bytes allocated for heap objects.",
//                 "record_value": 6.252166748046875,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "Sys",
//                 "record_description": "Total bytes of memory obtained from the OS.",
//                 "record_value": 13.844978332519531,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "Lookups",
//                 "record_description": "Number of pointer lookups performed by the runtime.",
//                 "record_value": 0,
//                 "record_unit": "bytes"
//             },
//             {
//                 "record_name": "Mallocs",
//                 "record_description": "Cumulative count of heap objects allocated.",
//                 "record_value": 2.572265625,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "Frees",
//                 "record_description": "Cumulative count of heap objects freed.",
//                 "record_value": 1.1806640625,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "HeapAlloc",
//                 "record_description": "Bytes of allocated heap objects.",
//                 "record_value": 5.075103759765625,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "HeapSys",
//                 "record_description": "Bytes of heap memory obtained from the OS.",
//                 "record_value": 7.46875,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "HeapIdle",
//                 "record_description": "Bytes in idle (unused) spans.",
//                 "record_value": 1.2890625,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "HeapInuse",
//                 "record_description": "Bytes in in-use spans.",
//                 "record_value": 6.1796875,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "HeapReleased",
//                 "record_description": "Bytes of physical memory returned to the OS.",
//                 "record_value": 592,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "HeapObjects",
//                 "record_description": "Number of allocated heap objects.",
//                 "record_value": 1.3916015625,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "StackInuse",
//                 "record_description": "Bytes in stack spans.",
//                 "record_value": 544,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "StackSys",
//                 "record_description": "Bytes of stack memory obtained from the OS.",
//                 "record_value": 544,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "MSpanInuse",
//                 "record_description": "Bytes of allocated mspan structures.",
//                 "record_value": 66.4453125,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "MSpanSys",
//                 "record_description": "Bytes of memory obtained from the OS for mspan structures.",
//                 "record_value": 79.5703125,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "MCacheInuse",
//                 "record_description": "Bytes of allocated mcache structures.",
//                 "record_value": 11.71875,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "MCacheSys",
//                 "record_description": "Bytes of memory obtained from the OS for mcache structures.",
//                 "record_value": 15.234375,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "BuckHashSys",
//                 "record_description": "Bytes of memory in profiling bucket hash tables.",
//                 "record_value": 1.379227638244629,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "GCSys",
//                 "record_description": "Bytes of memory in garbage collection metadata.",
//                 "record_value": 3.2230758666992188,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "OtherSys",
//                 "record_description": "Bytes of memory in miscellaneous off-heap runtime allocations.",
//                 "record_value": 1.1500921249389648,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "NextGC",
//                 "record_description": "Target heap size of the next GC cycle.",
//                 "record_value": 8.141136169433594,
//                 "record_unit": "MB"
//             },
//             {
//                 "record_name": "LastGC",
//                 "record_description": "Time the last garbage collection finished (nanoseconds since the UNIX epoch).",
//                 "record_value": 1607018250.3392446,
//                 "record_unit": "GB"
//             },
//             {
//                 "record_name": "PauseTotalNs",
//                 "record_description": "Cumulative nanoseconds in GC stop-the-world pauses since program start.",
//                 "record_value": 60.302734375,
//                 "record_unit": "KB"
//             },
//             {
//                 "record_name": "NumGC",
//                 "record_description": "Number of completed GC cycles.",
//                 "record_value": 1,
//                 "record_unit": "bytes"
//             },
//             {
//                 "record_name": "NumForcedGC",
//                 "record_description": "Number of GC cycles that were forced by the application calling GC.",
//                 "record_value": 0,
//                 "record_unit": "bytes"
//             },
//             {
//                 "record_name": "GCCPUFraction",
//                 "record_description": "Fraction of this program's available CPU time used by the GC.",
//                 "record_value": 0.00010095309082202974,
//                 "record_unit": "fraction"
//             }
//         ]
//     },
//     "heap_alloc_by_service": "5.09 MB",
//     "heap_alloc_by_system": "7.47 MB",
//     "total_alloc_by_service": "6.27 MB",
//     "total_memory_by_os": "13.84 MB",
//     "network_io": {
//         "bytes_sent": 2151682160,
//         "bytes_received": 2394602095
//     },
//     "overall_health": {
//         "overall_health_percent": "92.00%",
//         "health": {
//             "healthy": true,
//             "message": "[Healthy] Service is healthy and running smoothly."
//         }
//     }
// }

func StoreNewServiceMetrics(serviceMetrics *models.NewServiceStats) error {

	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	var rows []tstorage.Row
	timestamp := time.Now().Unix()

	log.Println("time.Now() ", time.Now())
	log.Println("timestamp Unix()", timestamp)

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
