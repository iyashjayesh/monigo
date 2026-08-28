package core

import (
	"strings"
	"sync"
	"testing"

	"github.com/iyashjayesh/monigo/models"
)

// breachStats returns a ServiceStats that, combined with the near-zero thresholds
// the breach tests configure, pushes every usage ratio past 100% regardless of
// what the host process is actually doing.
func breachStats() *models.ServiceStats {
	stats := &models.ServiceStats{}
	stats.CPUStatistics.TotalCores = 1
	stats.MemoryStatistics.TotalSystemMemory = "100 MB"
	stats.MemoryStatistics.MemoryUsedByService = "99 MB"
	stats.MemoryStatistics.MemoryUsedBySystem = "99 MB"
	return stats
}

// healthyStats returns a ServiceStats whose denominators are large enough that
// the live process CPU reading cannot dominate the score. Service CPU usage is
// divided by TotalCores, so a small core count makes the result depend on how
// busy the machine running the tests happens to be.
func healthyStats() *models.ServiceStats {
	stats := &models.ServiceStats{}
	stats.CPUStatistics.TotalCores = 100000
	stats.MemoryStatistics.TotalSystemMemory = "100000 MB"
	stats.MemoryStatistics.MemoryUsedByService = "1 MB"
	stats.MemoryStatistics.MemoryUsedBySystem = "1 MB"
	return stats
}

// withThresholds swaps in thresholds for the duration of fn and restores the
// originals afterwards, so tests do not leak configuration into one another.
func withThresholds(t *testing.T, th models.ServiceHealthThresholds, fn func()) {
	t.Helper()
	original := GetThresholds()
	ConfigureServiceThresholds(&th)
	defer ConfigureServiceThresholds(&original)
	fn()
}

// A service over every threshold must score 0, not 100. The breach branch used to
// clamp the score to 100 -- the best possible value -- so a service exceeding all
// of its limits reported as perfectly healthy.
func TestCalculateServiceHealthScoresZeroOnBreach(t *testing.T) {
	withThresholds(t, models.ServiceHealthThresholds{
		MaxCPUUsage:    0.0001,
		MaxMemoryUsage: 0.0001,
		MaxGoRoutines:  1,
	}, func() {
		score, msg, err := calculateServiceHealth(breachStats())
		if err != nil {
			t.Fatalf("calculateServiceHealth returned error: %v", err)
		}
		if score != 0 {
			t.Errorf("expected score 0 when usage exceeds all thresholds, got %v", score)
		}
		if msg == "" {
			t.Error("expected a non-empty breach message")
		}
	})
}

// Service and system health must agree on what a breach scores.
func TestServiceAndSystemHealthAgreeOnBreach(t *testing.T) {
	withThresholds(t, models.ServiceHealthThresholds{
		MaxCPUUsage:    0.0001,
		MaxMemoryUsage: 0.0001,
		MaxGoRoutines:  1,
	}, func() {
		serviceScore, _, err := calculateServiceHealth(breachStats())
		if err != nil {
			t.Fatalf("calculateServiceHealth returned error: %v", err)
		}
		systemScore, _, err := calculateSystemHealth(breachStats())
		if err != nil {
			t.Fatalf("calculateSystemHealth returned error: %v", err)
		}
		if serviceScore != systemScore {
			t.Errorf("service and system health disagree on breach: service=%v system=%v", serviceScore, systemScore)
		}
	})
}

// Usage comfortably inside the thresholds must not report a breach.
func TestCalculateServiceHealthWithinLimits(t *testing.T) {
	withThresholds(t, models.ServiceHealthThresholds{
		MaxCPUUsage:    100,
		MaxMemoryUsage: 100,
		MaxGoRoutines:  1000000,
	}, func() {
		score, msg, err := calculateServiceHealth(healthyStats())
		if err != nil {
			t.Fatalf("calculateServiceHealth returned error: %v", err)
		}
		if score <= 0 || score > 100 {
			t.Errorf("expected a score in (0, 100] when within limits, got %v", score)
		}
		if !strings.Contains(msg, "within limits") {
			t.Errorf("expected a within-limits message, got %q", msg)
		}
	})
}

func TestGetThresholdsReturnsConfiguredValues(t *testing.T) {
	withThresholds(t, models.ServiceHealthThresholds{
		MaxCPUUsage:    42,
		MaxMemoryUsage: 43,
		MaxGoRoutines:  44,
	}, func() {
		got := GetThresholds()
		if got.MaxCPUUsage != 42 || got.MaxMemoryUsage != 43 || got.MaxGoRoutines != 44 {
			t.Errorf("GetThresholds returned %+v, want {42 43 44}", got)
		}
	})
}

// Thresholds are written at configuration time and read from the metrics
// collection goroutine. Run under -race to catch unsynchronised access.
func TestThresholdsConcurrentAccess(t *testing.T) {
	original := GetThresholds()
	defer ConfigureServiceThresholds(&original)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			ConfigureServiceThresholds(&models.ServiceHealthThresholds{
				MaxCPUUsage:    float64(n%100 + 1),
				MaxMemoryUsage: float64(n%100 + 1),
				MaxGoRoutines:  n + 1,
			})
		}(i)
		go func() {
			defer wg.Done()
			_ = GetThresholds()
		}()
	}
	wg.Wait()
}

// Malformed memory strings must surface as errors, not panic. A short or empty
// value used to slice out of range inside common.ConvertToMB.
func TestCalculateMemoryUsagePercentageMalformedInput(t *testing.T) {
	for _, tc := range []struct{ used, total string }{
		{"", "100 MB"},
		{"M", "100 MB"},
		{"99 MB", ""},
		{"99 MB", "0 MB"},
		{"99 XB", "100 MB"},
	} {
		if _, err := calculateMemoryUsagePercentage(tc.used, tc.total); err == nil {
			t.Errorf("expected error for used=%q total=%q, got nil", tc.used, tc.total)
		}
	}
}
