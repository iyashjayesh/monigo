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
	ProcessId        int32     `json:"process_id"`
}

type GetMetrics struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type ServiceHealthThresholds struct {
	MaxGoroutines Thresholds `json:"max_goroutines"`
	MaxCPULoad    Thresholds `json:"max_cpu_load"`
	MaxMemory     Thresholds `json:"max_memory"`
}

type Thresholds struct {
	Value  float64 `json:"value"`
	Weight float64 `json:"weight"`
}
