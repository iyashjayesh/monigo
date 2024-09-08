package models

import (
	"time"

	"github.com/nakabonne/tstorage"
)

// ServiceInfo is the struct to store the service information
type ServiceInfo struct {
	ServiceName      string    `json:"service_name"`
	ServiceStartTime time.Time `json:"service_start_time"`
	GoVersion        string    `json:"go_version"`
	ProcessId        int32     `json:"process_id"`
}

// GetMetrics is the struct to get the metrics from the storage
type GetMetrics struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ServiceHealthThresholds is the struct to store the service health thresholds
type ServiceHealthThresholds struct {
	MaxGoroutines Thresholds `json:"max_goroutines"`
	MaxCPULoad    Thresholds `json:"max_cpu_load"`
	MaxMemory     Thresholds `json:"max_memory"`
}

// Thresholds is the struct to store the threshold values
type Thresholds struct {
	Value  float64 `json:"value"`
	Weight float64 `json:"weight"`
}

// FetchDataPoints is the struct to fetch the data points from the storage
type FetchDataPoints struct {
	FieldName []string `json:"field_name"`
	StartTime string   `json:"start_time"` // "2006-01-02T15:04:05Z07:00"
	EndTime   string   `json:"end_time"`   // "2006-01-02T15:04:05Z07:00"
}

// DataPointsInfo is the struct to store the data points information
type DataPointsInfo struct {
	FieldName string                `json:"field_name"`
	Data      []*tstorage.DataPoint `json:"data_points"`
}

// ReportsRequest is the struct to store the reports request
type ReportsRequest struct {
	Topic     string `json:"topic"`
	StartTime string `json:"start_time"` // "2006-01-02T15:04:05Z07:00"
	EndTime   string `json:"end_time"`   // "2006-01-02T15:04:05Z07:00"
	TimeFrame string `json:"time_frame"`
}
