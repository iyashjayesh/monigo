package models

import (
	"time"
)

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

// db struct
type ServiceInfo struct {
	ServiceName      string
	ServiceStartTime time.Time
	GoVersion        string
	TimeStamp        time.Time
}

type ServiceMetrics struct {
	Load                   float64
	Cores                  float64
	MemoryUsed             float64
	NumberOfReqServerd     float64
	GoRoutines             float64
	TotalAlloc             float64
	MemoryAllocSys         float64
	HeapAlloc              float64
	HeapAllocSys           float64
	TotalDurationTookByAPI time.Duration
}

type GetMetrics struct {
	Name  string
	Start time.Time
	End   time.Time
}
