package models

// ServiceCoreStats represents the core statistics of the service.
type ServiceCoreStats struct {
	Goroutines                 int        `json:"goroutines"`
	Requests                   int64      `json:"requests"`
	Load                       string     `json:"load"`
	CPU                        CPUStat    `json:"cpu"`
	Memory                     MemoryStat `json:"memory"`
	Uptime                     string     `json:"uptime"`
	TotalDurationTookbyRequest float64    `json:"total_duration_took_by_request"`
}

// CPUStat represents the CPU statistics.
type CPUStat struct {
	TotalCores        float64 `json:"total_cores"`
	TotalLogicalCores float64 `json:"total_logical_cores"`
	SystemUsedCores   float64 `json:"system_used_cores"`
	ProcessUsedCores  float64 `json:"process_used_cores"`
	Cores             string  `json:"cores"`
	UsedInPercent     string  `json:"used_in_percent"` // by process
}

// MemoryStat represents the memory statistics.
type MemoryStat struct {
	TotalMemory         float64         `json:"total_memory"`
	UsedBySystem        float64         `json:"used_by_system"`
	UsedByProcess       float64         `json:"used_by_process"`
	FreeMemory          float64         `json:"free_memory"`
	HeapAllocByProcess  float64         `json:"heap_alloc_by_process"`  // HeapAlloc -> heap alloc means heap allocated by application
	HeapSysByProcess    float64         `json:"heap_sys_by_process"`    // HeapSys -> heap sys means heap allocated by system
	TotalAllocByProcess float64         `json:"total_alloc_by_process"` // TotalAlloc is the total allocated bytes in the lifetime of the process. Every allocation is counted.
	TotalSysByProcess   float64         `json:"total_sys_by_process"`   // Sys is the total bytes of memory obtained from the OS.
	UsedInPercent       string          `json:"used_in_percent"`        // UsedInPercent is the percentage of memory used by the process.
	MemStatsRecords     MemStatsRecords `json:"mem_stats_records"`      // MemStatsRecords holds a list of memory statistic records.
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
