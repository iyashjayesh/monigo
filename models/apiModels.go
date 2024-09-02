package models

import "time"

// NewServiceStats represents the final statistics of the service.
type NewServiceStats struct {
	CoreStatistics   CoreStatistics   `json:"core_statistics"`   // Core Statistics
	LoadStatistics   LoadStatistics   `json:"load_statistics"`   // Load Statistics
	CPUStatistics    CPUStatistics    `json:"cpu_statistics"`    // CPU Statistics
	MemoryStatistics MemoryStatistics `json:"memory_statistics"` // Memory Statistics

	// Additional Metrics
	HeapAllocByService  float64 `json:"heap_alloc_by_service"`
	HeapAllocBySystem   float64 `json:"heap_alloc_by_system"`
	TotalAllocByService float64 `json:"total_alloc_by_service"`
	TotalMemoryByOS     float64 `json:"total_memory_by_os"`
	// DiskIO            float64 `json:"disk_io"` @TODO: Need to work on this
	NetworkIO struct {
		BytesSent     float64 `json:"bytes_sent"`
		BytesReceived float64 `json:"bytes_received"`
	} `json:"network_io"`

	// Health
	OverallHealth ServiceHealth `json:"overall_health"`
}

type CoreStatistics struct {
	Goroutines                 int           `json:"goroutines"`
	RequestCount               int64         `json:"request_count"`
	Uptime                     string        `json:"uptime"`
	TotalDurationTookByRequest time.Duration `json:"total_duration_took_by_request"`
}

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

type CPUStatistics struct {
	TotalCores                  float64 `json:"total_cores"`
	TotalLogicalCores           float64 `json:"total_logical_cores"`
	CoresUsedBySystem           float64 `json:"cores_used_by_system"`
	CoresUsedByService          float64 `json:"cores_used_by_service"`
	CoresUsedByServiceInPercent string  `json:"cores_used_by_service_in_percent"`
	CoresUsedBySystemInPercent  string  `json:"cores_used_by_system_in_percent"`
}

type MemoryStatistics struct {
	TotalSystemMemory   float64  `json:"total_system_memory"`
	MemoryUsedBySystem  float64  `json:"memory_used_by_system"`
	MemoryUsedByService float64  `json:"memory_used_by_service"`
	AvailableMemory     float64  `json:"available_memory"`
	GCPauseDuration     float64  `json:"gc_pause_duration"`
	StackMemoryUsage    float64  `json:"stack_memory_usage"`
	TotalSwapMemory     float64  `json:"total_swap_memory"`
	FreeSwapMemory      float64  `json:"free_swap_memory"`
	MemStatsRecords     []Record `json:"mem_stats_records"` // List of memory statistic records.
}

////////////////////

// ServiceCoreStats represents the core statistics of the service.
type ServiceCoreStats struct {
	Goroutines                 int        `json:"goroutines"`
	RequestCount               int64      `json:"request_count"`
	Loads                      Loads      `json:"loads"`
	CPU                        CPUStat    `json:"cpu"`
	Memory                     MemoryStat `json:"memory"`
	Uptime                     string     `json:"uptime"`
	TotalDurationTookbyRequest float64    `json:"total_duration_took_by_request"`
}

// Loads represents the load statistics.
type Loads struct {
	ServiceCPULoad       string `json:"service_cpu_load"`
	SystemCPULoad        string `json:"system_cpu_load"`
	TotalCPULoad         string `json:"total_cpu_load"`
	ServiceMemLoad       string `json:"service_memory_load"`
	SystemMemLoad        string `json:"system_memory_load"`
	TotalMemLoad         string `json:"total_memory_load"`
	ServcieDiskLoad      string `json:"service_disk_load"`
	SystemDiskLoad       string `json:"system_disk_load"`
	TotalDiskLoad        string `json:"total_disk_load"`
	OverallLoadOfService string `json:"overall_load_of_service"` // Final load of the service
}

// CPUStat represents the CPU statistics.
type CPUStat struct {
	TotalCores                  float64 `json:"total_cores"`
	TotalLogicalCores           float64 `json:"total_logical_cores"`
	CoresUsedBySystem           float64 `json:"cores_used_by_system"`
	CoresUsedByService          float64 `json:"cores_used_by_service"`
	CoresUsedByServiceInPercent string  `json:"cores_used_by_service_in_percent"`
	CoresUsedBySystemInPercent  string  `json:"cores_used_by_system_in_percent"`
}

// MemoryStat represents the memory statistics.
type MemoryStat struct {
	TotalSystemMemory   float64         `json:"total_system_memory"`
	MemoryUsedBySystem  float64         `json:"memory_used_by_system"`
	MemoryUsedByService float64         `json:"memory_used_by_service"`
	AvaialbleMemory     float64         `json:"available_memory"`
	MemStatsRecords     MemStatsRecords `json:"mem_stats_records"` // MemStatsRecords holds a list of memory statistic records.
	// HeapAllocByProcess  float64         `json:"heap_alloc_by_process"`  // HeapAlloc -> heap alloc means heap allocated by application
	// HeapSysByProcess    float64         `json:"heap_sys_by_process"`    // HeapSys -> heap sys means heap allocated by system
	// TotalAllocByProcess float64         `json:"total_alloc_by_process"` // TotalAlloc is the total allocated bytes in the lifetime of the process. Every allocation is counted.
	// TotalSysByProcess   float64         `json:"total_sys_by_process"`   // Sys is the total bytes of memory obtained from the OS.
	// UsedInPercent       string          `json:"used_in_percent"`        // UsedInPercent is the percentage of memory used by the process.
}

// MemStatsRecords holds a list of memory statistic records.
type MemStatsRecords struct {
	Records []Record `json:"records"`
}

// Record represents a single memory statistic record.
type Record struct {
	Name        string      `json:"record_name"`
	Description string      `json:"record_description"`
	Value       interface{} `json:"record_value"`
	Unit        string      `json:"record_unit,omitempty"` // Added Unit to support different units like bytes, MB, GB, etc.
}

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

type ServiceHealth struct {
	OverallHealthPercent string `json:"overall_health_percent"`
	Health               Health `json:"health"`
}

type Health struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

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