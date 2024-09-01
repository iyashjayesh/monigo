package models

import (
	"time"
)

type ProcessStats struct {
	ProcessId         int32   `json:"process_id"`
	SysCPUPercent     float64 `json:"sys_cpu_percent"`
	ProcCPUPercent    float64 `json:"proc_cpu_percent"`
	ProcMemPercent    float64 `json:"proc_mem_percent"`
	TotalMemoryUsage  float64 `json:"total_memory_usage"`
	FreeMemory        float64 `json:"free_memory"`
	UsedMemoryPercent float64 `json:"used_memory_percent"`
	TotalCores        int     `json:"total_cores"`
	TotalLogicalCores int     `json:"total_logical_cores"`
	SystemUsedCores   float64 `json:"system_used_cores"`
	ProcessUsedCores  float64 `json:"process_used_cores"`
}

type Memory struct {
	TotalMemoryUsage float64 `json:"total_memory_usage"`
	FreeMemory       float64 `json:"free_memory"`
	UsedPercent      float64 `json:"used_percent"`
}

type FunctionMetrics struct {
	FunctionLastRanAt time.Time     `json:"function_last_ran_at"`
	CPUProfile        string        `json:"cpu_profile"`
	MemoryUsage       uint64        `json:"memory_usage"`
	GoroutineCount    int           `json:"goroutine_count"`
	ExecutionTime     time.Duration `json:"execution_time"`
}

type ServiceInfo struct {
	ServiceName      string    `json:"service_name"`
	ServiceStartTime time.Time `json:"service_start_time"`
	GoVersion        string    `json:"go_version"`
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

type GetMetrics struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ServiceHealth struct {
	Goroutines           int     `json:"goroutines"`
	Requests             int     `json:"requests"`
	MemoryUsed           float64 `json:"memory_used"`
	CPUPercent           float64 `json:"cpu_percent"`
	OverallHealthPercent float64 `json:"overall_health_percent"`
	Health               Health  `json:"health"`
}

type Health struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

type ServiceHealthThresholds struct {
	MaxGoroutines Thresholds `json:"max_goroutines"`
	MaxLoad       Thresholds `json:"max_load"`
	MaxMemory     Thresholds `json:"max_memory"`
}

type Thresholds struct {
	Value  float64 `json:"value"`
	Weight float64 `json:"weight"`
}
