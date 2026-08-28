package models

import (
	"time"
)

// ServiceInfo is the struct to store the service information
type ServiceInfo struct {
	ServiceName      string    `json:"service_name"`
	ServiceStartTime time.Time `json:"service_start_time"`
	GoVersion        string    `json:"go_version"`
	ProcessId        int32     `json:"process_id"`

	// Storage footprint. Retention and where the data lives are properties of
	// the instrument rather than of any one metric, so the dashboard shows them
	// in its chrome. Both are omitted when unset rather than reported as zero,
	// since "0s retention" and "not configured" are different states.
	RetentionPeriod string `json:"retention_period,omitempty"`
	StorageType     string `json:"storage_type,omitempty"`
	StorageOnDisk   string `json:"storage_on_disk,omitempty"`
}

// ServiceHealthThresholds is the struct to store the service health thresholds
type ServiceHealthThresholds struct {
	MaxCPUUsage    float64 `json:"max_cpu_usage"`    // Default is 80%
	MaxMemoryUsage float64 `json:"max_memory_usage"` // Default is 80%
	MaxGoRoutines  int     `json:"max_go_routines"`  // Default is 1000
}

// FetchDataPoints is the struct to fetch the data points from the storage
type FetchDataPoints struct {
	FieldName []string `json:"field_name"`
	StartTime string   `json:"start_time"` // "2006-01-02T15:04:05Z07:00"
	EndTime   string   `json:"end_time"`   // "2006-01-02T15:04:05Z07:00"
}

// DataPointsInfo is the struct to store the data points information
type DataPointsInfo struct {
	FieldName string        `json:"field_name"`
	Data      []interface{} `json:"data_points"`
}

// ReportsRequest is the struct to store the reports request
type ReportsRequest struct {
	Topic     string `json:"topic"`
	StartTime string `json:"start_time"` // "2006-01-02T15:04:05Z07:00"
	EndTime   string `json:"end_time"`   // "2006-01-02T15:04:05Z07:00"
	TimeFrame string `json:"time_frame"`
}

// SystemHealthInPercent is the struct to store the system health in percentage
type SystemHealthInPercent struct {
	SystemHealth  HealthFields `json:"system_health_percentage"`
	ServiceHealth HealthFields `json:"service_health_percentage"`
}

type HealthFields struct {
	Percentage    float64 `json:"percentage"`
	AllowedByUser float64 `json:"allowed_by_user"`
	Message       string  `json:"message"`
}
