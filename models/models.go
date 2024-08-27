package models

import (
	"time"

	"github.com/google/uuid"
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
	GoVerison        string
	TimeStamp        time.Time
}

type ServiceMetrics struct {
	Id                     uuid.UUID
	Load                   string
	Cores                  string
	MemoryUsed             string
	UpTime                 time.Duration
	NumberOfReqServerd     int64
	TotalDurationTookByAPI time.Duration
	GoRoutines             int64
	TotalAlloc             uint64
	MemoryAllocSys         uint64
	HeapAlloc              uint64
	HeapAllocSys           uint64
	TimeStamp              time.Time
}
