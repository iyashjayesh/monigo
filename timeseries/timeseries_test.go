package timeseries

import (
	"runtime"
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

func init() {
	common.SetServiceInfo("test-service", time.Now(), runtime.Version(), 1234, "7d")
}

func TestInMemoryStorage_InsertAndSelect(t *testing.T) {
	s := NewInMemoryStorage()

	now := time.Now().Unix()
	rows := []Row{
		{Metric: "cpu_load", DataPoint: DataPoint{Timestamp: now, Value: 45.5}, Labels: []Label{{Name: "host", Value: "test"}}},
		{Metric: "cpu_load", DataPoint: DataPoint{Timestamp: now + 10, Value: 55.0}, Labels: []Label{{Name: "host", Value: "test"}}},
		{Metric: "mem_load", DataPoint: DataPoint{Timestamp: now, Value: 70.0}, Labels: []Label{{Name: "host", Value: "test"}}},
	}

	if err := s.InsertRows(rows); err != nil {
		t.Fatalf("InsertRows error: %v", err)
	}

	// Select cpu_load
	points, err := s.Select("cpu_load", nil, now-1, now+20)
	if err != nil {
		t.Fatalf("Select error: %v", err)
	}
	if len(points) != 2 {
		t.Errorf("expected 2 cpu_load points, got %d", len(points))
	}

	// Select with time range filter
	points, err = s.Select("cpu_load", nil, now+5, now+20)
	if err != nil {
		t.Fatalf("Select error: %v", err)
	}
	if len(points) != 1 {
		t.Errorf("expected 1 cpu_load point in range, got %d", len(points))
	}

	// Select non-existent metric
	points, err = s.Select("nonexistent", nil, now-1, now+20)
	if err != nil {
		t.Fatalf("Select error: %v", err)
	}
	if points != nil {
		t.Errorf("expected nil for nonexistent metric, got %v", points)
	}
}

func TestInMemoryStorage_Close(t *testing.T) {
	s := NewInMemoryStorage()
	if err := s.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestGetHostLabel(t *testing.T) {
	label := GetHostLabel()
	if label.Name != "host" {
		t.Errorf("expected label name 'host', got %q", label.Name)
	}
	if label.Value == "" {
		t.Error("expected non-empty hostname value")
	}
}

func TestStoreAndRetrieveMetrics(t *testing.T) {
	// Use in-memory storage for tests
	SetStorageType("memory")
	manager = &storageManager{} // Reset singleton

	_, err := GetStorageInstance()
	if err != nil {
		t.Fatalf("GetStorageInstance error: %v", err)
	}

	stats := models.ServiceStats{
		CoreStatistics: models.CoreStatistics{Goroutines: 10},
		LoadStatistics: models.LoadStatistics{
			ServiceCPULoadRaw:       25.0,
			SystemCPULoadRaw:        40.0,
			ServiceMemLoadRaw:       30.0,
			SystemMemLoadRaw:        60.0,
			OverallLoadOfServiceRaw: 27.5,
			SystemDiskLoadRaw:       50.0,
			TotalDiskLoadRaw:        100.0,
		},
		CPUStatistics: models.CPUStatistics{
			TotalCores:         8,
			CoresUsedByService: 2,
			CoresUsedBySystem:  4,
		},
		MemoryStatistics: models.MemoryStatistics{
			TotalSystemMemoryRaw:   16000000000,
			MemoryUsedBySystemRaw:  8000000000,
			MemoryUsedByServiceRaw: 500000000,
			AvailableMemoryRaw:     8000000000,
			GCPauseDurationRaw:     1.5,
			StackMemoryUsageRaw:    1000000,
		},
		NetworkIO: struct {
			BytesSent     float64 `json:"bytes_sent"`
			BytesReceived float64 `json:"bytes_received"`
		}{BytesSent: 1000, BytesReceived: 2000},
		Health: models.ServiceHealth{
			ServiceHealth: models.Health{Percent: 85},
			SystemHealth:  models.Health{Percent: 90},
		},
	}

	if err := StoreServiceMetrics(&stats); err != nil {
		t.Fatalf("StoreServiceMetrics error: %v", err)
	}

	// Retrieve goroutines metric
	label := GetHostLabel()
	now := time.Now().Unix()
	points, err := GetDataPoints("goroutines", []Label{label}, now-10, now+10)
	if err != nil {
		t.Fatalf("GetDataPoints error: %v", err)
	}
	if len(points) == 0 {
		t.Error("expected at least 1 goroutines data point")
	}
	if points[0].Value != 10 {
		t.Errorf("expected goroutines value 10, got %f", points[0].Value)
	}

	// Retrieve CPU metric
	points, err = GetDataPoints("service_cpu_load", []Label{label}, now-10, now+10)
	if err != nil {
		t.Fatalf("GetDataPoints error: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("expected at least 1 service_cpu_load data point")
	}
	if points[0].Value != 25.0 {
		t.Errorf("expected service_cpu_load 25.0, got %f", points[0].Value)
	}

	// Cleanup
	CloseStorage()
}
