package monigo

import "time"

type ProcessStats struct {
	ProcessId         int32
	SysCPUPercent     float64
	ProcCPUPercent    float64
	ProcMemPercent    float64
	TotalMemoryUsage  float64
	FreeMemory        float64
	UsedMemoryPercent float64
	TotalCores        int
	TotalLogicalCores int
	SystemUsedCores   float64
	ProcessUsedCores  float64
}

type Memory struct {
	TotalMemoryUsage float64
	FreeMemory       float64
	UsedPercent      float64
}

type FunctionMetrics struct {
	FunctionLastRanAt time.Time
	CPUProfile        string
	MemoryUsage       uint64
	GoroutineCount    int
	ExecutionTime     time.Duration
}
