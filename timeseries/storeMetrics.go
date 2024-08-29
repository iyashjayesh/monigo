package timeseries

import (
	"fmt"
	"log"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/nakabonne/tstorage"
)

// StoreServiceMetrics stores the service metrics in the storage.
func StoreServiceMetrics(serviceMetrics *models.ServiceMetrics) error {
	sto, err := GetStorageInstance()
	if err != nil {
		return fmt.Errorf("error getting storage instance: %w", err)
	}

	var rows []tstorage.Row
	timestamp := time.Now().Unix()
	rows = []tstorage.Row{
		{
			Metric:    "load_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Load},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "cores_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.Cores},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "memory_used_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryUsed},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "number_of_req_served_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.NumberOfReqServerd},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "goroutines_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.GoRoutines},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "total_alloc_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.TotalAlloc},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "memory_alloc_sys_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.MemoryAllocSys},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "heap_alloc_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.HeapAlloc},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "heap_alloc_sys_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: serviceMetrics.HeapAllocSys},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
		{
			Metric:    "total_duration_metrics",
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: float64(serviceMetrics.TotalDurationTookByAPI.Seconds())},
			Labels:    []tstorage.Label{{Name: "host", Value: "server1"}},
		},
	}

	if err := sto.InsertRows(rows); err != nil {
		return fmt.Errorf("error storing service metrics: %w", err)
	}

	log.Println("Stored service metrics, timestamp:", timestamp)
	return nil
}

// GetDataPoints retrieves data points for a given metric and labels.
func GetDataPoints(metric string, labels []tstorage.Label, start, end int64) ([]*tstorage.DataPoint, error) {
	sto, err := GetStorageInstance()
	if err != nil {
		return nil, fmt.Errorf("error getting storage instance: %w", err)
	}
	return sto.Select(metric, labels, start, end)
}
