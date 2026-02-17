package timeseries

import (
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

func BenchmarkInMemoryStorageInsert(b *testing.B) {
	s := NewInMemoryStorage()
	now := time.Now().Unix()
	rows := []Row{
		{Metric: "bench_metric", DataPoint: DataPoint{Timestamp: now, Value: 42.0}, Labels: []Label{{Name: "host", Value: "bench"}}},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows[0].DataPoint.Timestamp = now + int64(i)
		s.InsertRows(rows)
	}
}

func BenchmarkInMemoryStorageSelect(b *testing.B) {
	s := NewInMemoryStorage()
	now := time.Now().Unix()
	for i := 0; i < 1000; i++ {
		s.InsertRows([]Row{
			{Metric: "bench_metric", DataPoint: DataPoint{Timestamp: now + int64(i), Value: float64(i)}, Labels: []Label{{Name: "host", Value: "bench"}}},
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Select("bench_metric", nil, now, now+1000)
	}
}

func BenchmarkStoreServiceMetrics(b *testing.B) {
	SetStorageType("memory")
	manager = &storageManager{}
	GetStorageInstance()

	stats := &models.ServiceStats{
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
		CPUStatistics: models.CPUStatistics{TotalCores: 8, CoresUsedByService: 2, CoresUsedBySystem: 4},
		MemoryStatistics: models.MemoryStatistics{
			TotalSystemMemoryRaw:   16e9,
			MemoryUsedBySystemRaw:  8e9,
			MemoryUsedByServiceRaw: 5e8,
			AvailableMemoryRaw:     8e9,
			GCPauseDurationRaw:     1.5,
			StackMemoryUsageRaw:    1e6,
		},
		Health: models.ServiceHealth{
			ServiceHealth: models.Health{Percent: 85},
			SystemHealth:  models.Health{Percent: 90},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StoreServiceMetrics(stats)
	}
}
