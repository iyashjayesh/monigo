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
	Low            float64 `json:"low"`
	Medium         float64 `json:"medium"`
	High           float64 `json:"high"`
	Critical       float64 `json:"critical"`
	GoRoutinesLow  int     `json:"go_routines_low"`
	GoRoutinesHigh int     `json:"go_routines_high"`
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

// SystemHealthInPercent is the struct to store the system health in percentage
type SystemHealthInPercent struct {
	SystemHealth  float64 `json:"system_health_percentage"`
	ServiceHealth float64 `json:"service_health_percentage"`
	OverallHealth float64 `json:"overall_health_percentage"`
}
