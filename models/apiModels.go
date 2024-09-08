package models

import (
	"time"
)

// NewServiceStats represents the final statistics of the service.
type NewServiceStats struct {
	CoreStatistics   CoreStatistics   `json:"core_statistics"`   // Core Statistics
	LoadStatistics   LoadStatistics   `json:"load_statistics"`   // Load Statistics
	CPUStatistics    CPUStatistics    `json:"cpu_statistics"`    // CPU Statistics
	MemoryStatistics MemoryStatistics `json:"memory_statistics"` // Memory Statistics

	// Additional Metrics
	HeapAllocByService  string `json:"heap_alloc_by_service"`
	HeapAllocBySystem   string `json:"heap_alloc_by_system"`
	TotalAllocByService string `json:"total_alloc_by_service"`
	TotalMemoryByOS     string `json:"total_memory_by_os"`
	// DiskIO            float64 `json:"disk_io"` @TODO: Need to work on this
	NetworkIO struct {
		BytesSent     float64 `json:"bytes_sent"`
		BytesReceived float64 `json:"bytes_received"`
	} `json:"network_io"`

	// Health
	OverallHealth ServiceHealth `json:"overall_health"`
}

// CoreStatistics represents the core statistics of the service.
type CoreStatistics struct {
	Goroutines                 int           `json:"goroutines"`
	RequestCount               int64         `json:"request_count"`
	Uptime                     string        `json:"uptime"`
	TotalDurationTookByRequest time.Duration `json:"total_duration_took_by_request"`
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
	// ServiceDiskLoad      string `json:"service_disk_load"`  @TODO: Need to work on this
	// SystemDiskLoad       string `json:"system_disk_load"`   @TODO: Need to work on this
	// TotalDiskLoad        string `json:"total_disk_load"`	   @TODO: Need to work on this
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
}

// ServiceInfo is the struct to store the service information
type ServiceHealth struct {
	OverallHealthPercent string `json:"overall_health_percent"`
	Health               Health `json:"health"`
}

// Health represents the health of the service.
type Health struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
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

// ServiceMetrics represents the metrics of the service.
type ServiceMetrics struct {
	Load                   float64       `json:"load"`
	Cores                  float64       `json:"cores"`
	MemoryUsed             float64       `json:"memory_used"`
	NumberOfReqServerd     float64       `json:"number_of_req_served"`
	GoRoutines             float64       `json:"go_routines"`
	TotalAlloc             float64       `json:"total_alloc"`
	MemoryAllocSys         float64       `json:"memory_alloc_sys"`
	HeapAlloc              float64       `json:"heap_alloc"`
	HeapAllocSys           float64       `json:"heap_alloc_sys"`
	TotalDurationTookByAPI time.Duration `json:"total_duration_took_by_api"`
	OverallHealth          ServiceHealth `json:"overall_health"`
}

// FieldName represents the field names of the service.
type FieldName struct {
	HeapAllocByService           float64
	HeapAllocBySystem            float64
	TotalAllocByService          float64
	TotalMemoryByOS              float64
	MemoryUsedInPercentByService float64
	GCPauseDuration              float64
	NumberOfGoroutines           float64
	CPUUsageByService            float64
	StackMemoryUsage             float64
	TotalSwapMemory              float64
	FreeSwapMemory               float64
	DiskIO                       float64
	NetworkIO                    float64
}

// GoRoutinesStatistic represents the Go routines statistics.
type GoRoutinesStatistic struct {
	NumberOfGoroutines int      `json:"number_of_goroutines"`
	StackView          []string `json:"stack_view"`
}
