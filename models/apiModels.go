package models

import (
	"time"
)

// ServiceStats represents the final statistics of the service.
type ServiceStats struct {
	CoreStatistics   CoreStatistics   `json:"core_statistics"`   // Core Statistics
	LoadStatistics   LoadStatistics   `json:"load_statistics"`   // Load Statistics
	CPUStatistics    CPUStatistics    `json:"cpu_statistics"`    // CPU Statistics
	MemoryStatistics MemoryStatistics `json:"memory_statistics"` // Memory Statistics

	// Additional Metrics
	HeapAllocByService  string `json:"heap_alloc_by_service"`
	HeapAllocBySystem   string `json:"heap_alloc_by_system"`
	TotalAllocByService string `json:"total_alloc_by_service"`
	TotalMemoryByOS     string `json:"total_memory_by_os"`

	// Raw values for storage
	HeapAllocByServiceRaw  uint64 `json:"-"`
	HeapAllocBySystemRaw   uint64 `json:"-"`
	TotalAllocByServiceRaw uint64 `json:"-"`
	TotalMemoryByOSRaw     uint64 `json:"-"`

	DiskIO struct {
		ReadBytes  uint64 `json:"read_bytes"`
		WriteBytes uint64 `json:"write_bytes"`
	} `json:"disk_io"` // Disk Use percentage
	NetworkIO struct {
		BytesSent     float64 `json:"bytes_sent"`
		BytesReceived float64 `json:"bytes_received"`
	} `json:"network_io"`

	// Health
	Health ServiceHealth `json:"health"`
}

// CoreStatistics represents the core statistics of the service.
type CoreStatistics struct {
	Goroutines int    `json:"goroutines"`
	Uptime     string `json:"uptime"`
	// RequestCount               int64         `json:"request_count"`
	// TotalDurationTookByRequest time.Duration `json:"total_duration_took_by_request"`
}

// LoadStatistics represents the load statistics of the service.
type LoadStatistics struct {
	ServiceCPULoad       string `json:"service_cpu_load"`
	SystemCPULoad        string `json:"system_cpu_load"`
	TotalCPULoad         string `json:"total_cpu_load"`
	ServiceMemLoad       string `json:"service_memory_load"`
	SystemMemLoad        string `json:"system_memory_load"`
	TotalMemLoad         string `json:"total_memory_load"`
	OverallLoadOfService string `json:"overall_load_of_service"` // Final load of the service
	ServiceDiskLoad      string `json:"service_disk_load"`
	SystemDiskLoad       string `json:"system_disk_load"`
	TotalDiskLoad        string `json:"total_disk_load"`
	// Raw values for storage
	ServiceCPULoadRaw       float64 `json:"-"`
	SystemCPULoadRaw        float64 `json:"-"`
	ServiceMemLoadRaw       float64 `json:"-"`
	SystemMemLoadRaw        float64 `json:"-"`
	OverallLoadOfServiceRaw float64 `json:"-"`
	SystemDiskLoadRaw       float64 `json:"-"`
	TotalDiskLoadRaw        float64 `json:"-"`
}

// CPUStatistics represents the CPU statistics of the service.
type CPUStatistics struct {
	TotalCores                  float64 `json:"total_cores"`
	TotalLogicalCores           float64 `json:"total_logical_cores"`
	CoresUsedBySystem           float64 `json:"cores_used_by_system"`
	CoresUsedByService          float64 `json:"cores_used_by_service"`
	CoresUsedByServiceInPercent string  `json:"cores_used_by_service_in_percent"`
	CoresUsedBySystemInPercent  string  `json:"cores_used_by_system_in_percent"`
}

// MemoryStatistics represents the memory statistics of the service.
type MemoryStatistics struct {
	TotalSystemMemory   string               `json:"total_system_memory"`
	MemoryUsedBySystem  string               `json:"memory_used_by_system"`
	MemoryUsedByService string               `json:"memory_used_by_service"`
	AvailableMemory     string               `json:"available_memory"`
	GCPauseDuration     string               `json:"gc_pause_duration"`
	StackMemoryUsage    string               `json:"stack_memory_usage"`
	TotalSwapMemory     string               `json:"total_swap_memory"`
	FreeSwapMemory      string               `json:"free_swap_memory"`
	MemStatsRecords     []Record             `json:"mem_stats_records"`     // List of memory statistic records.
	RawMemStatsRecords  []RawMemStatsRecords `json:"raw_mem_stats_records"` // RawMemStatsRecords holds a list of raw memory statistic records.
	// Raw values for storage
	TotalSystemMemoryRaw   float64 `json:"-"`
	MemoryUsedBySystemRaw  float64 `json:"-"`
	MemoryUsedByServiceRaw float64 `json:"-"`
	AvailableMemoryRaw     float64 `json:"-"`
	GCPauseDurationRaw     float64 `json:"-"`
	StackMemoryUsageRaw    float64 `json:"-"`
}

// ServiceHealth represents the health of the service.
type ServiceHealth struct {
	SystemHealth  Health `json:"system_health"`
	ServiceHealth Health `json:"service_health"`
}

// Health represents the health of the service.
type Health struct {
	Percent float64 `json:"percent"`
	Healthy bool    `json:"healthy"`
	Message string  `json:"message"`
	IconMsg string  `json:"icon_msg"`
}

// RawMemStatsRecords holds a list of raw memory statistic records.
type RawMemStatsRecords struct {
	RecordName  string  `json:"record_name"`
	RecordValue float64 `json:"record_value"`
}

// Record represents a single memory statistic record.
type Record struct {
	Name        string      `json:"record_name"`
	Description string      `json:"record_description"`
	Value       interface{} `json:"record_value"`
	Unit        string      `json:"record_unit,omitempty"` // Added Unit to support different units like bytes, MB, GB, etc.
}

// GoRoutinesStatistic represents the Go routines statistics.
type GoRoutinesStatistic struct {
	NumberOfGoroutines int      `json:"number_of_goroutines"`
	StackView          []string `json:"stack_view"`
}

// FunctionTraceDetails represents the function trace details.
type FunctionTraceDetails struct {
	FunctionName      string   `json:"function_name"`
	CoreProfile       Profiles `json:"core_profile"`
	FunctionCodeTrace string   `json:"function_code_trace"`
}

// Profiles represents the profiles.
type Profiles struct {
	CPU string `json:"cpu_profile"`
	Mem string `json:"mem_profile"`
}

// FunctionMetrics represents the function metrics.
type FunctionMetrics struct {
	FunctionLastRanAt  time.Time     `json:"function_last_ran_at"`
	CPUProfileFilePath string        `json:"cpu_profile_file_path"`
	MemProfileFilePath string        `json:"mem_profile_file_path"`
	MemoryUsage        uint64        `json:"memory_usage"`
	GoroutineCount     int           `json:"goroutine_count"`
	ExecutionTime      time.Duration `json:"execution_time"`
}
