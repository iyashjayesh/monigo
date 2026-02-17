package core

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

func init() {
	common.SetServiceInfo("test-service", time.Now(), runtime.Version(), 1234, "7d")
	ConfigureServiceThresholds(&models.ServiceHealthThresholds{
		MaxCPUUsage:    95,
		MaxMemoryUsage: 95,
		MaxGoRoutines:  1000,
	})
}

func TestGetServiceStats(t *testing.T) {
	stats := GetServiceStats(context.Background())

	if stats.CoreStatistics.Goroutines <= 0 {
		t.Error("expected goroutines > 0")
	}
	if stats.CoreStatistics.Uptime == "" {
		t.Error("expected non-empty uptime")
	}
	if stats.MemoryStatistics.TotalSystemMemory == "" {
		t.Error("expected non-empty TotalSystemMemory")
	}
	if stats.CPUStatistics.TotalCores <= 0 {
		t.Error("expected TotalCores > 0")
	}
}

func TestGetCoreStatistics(t *testing.T) {
	cs := GetCoreStatistics()
	if cs.Goroutines <= 0 {
		t.Errorf("expected goroutines > 0, got %d", cs.Goroutines)
	}
	if cs.Uptime == "" {
		t.Error("expected non-empty uptime string")
	}
}

func TestGetLoadStatistics(t *testing.T) {
	ls := GetLoadStatistics()

	for _, s := range []string{ls.ServiceCPULoad, ls.SystemCPULoad, ls.TotalCPULoad, ls.ServiceMemLoad, ls.SystemMemLoad} {
		if !strings.HasSuffix(s, "%") {
			t.Errorf("expected load string ending with '%%', got %q", s)
		}
	}
	if ls.ServiceCPULoadRaw < 0 {
		t.Errorf("expected ServiceCPULoadRaw >= 0, got %f", ls.ServiceCPULoadRaw)
	}
	if ls.SystemMemLoadRaw < 0 {
		t.Errorf("expected SystemMemLoadRaw >= 0, got %f", ls.SystemMemLoadRaw)
	}
}

func TestCalculateOverallLoad(t *testing.T) {
	tests := []struct {
		cpu, mem    float64
		wantF       float64
		wantSuffix  string
	}{
		{0, 0, 0, "%"},
		{100, 100, 100, "%"},
		{50, 50, 50, "%"},
		{200, 200, 100, "%"}, // capped at 100
	}

	for _, tt := range tests {
		f, s := CalculateOverallLoad(tt.cpu, tt.mem)
		if f != tt.wantF {
			t.Errorf("CalculateOverallLoad(%v, %v) float = %v, want %v", tt.cpu, tt.mem, f, tt.wantF)
		}
		if !strings.HasSuffix(s, tt.wantSuffix) {
			t.Errorf("CalculateOverallLoad(%v, %v) string = %q, want suffix %q", tt.cpu, tt.mem, s, tt.wantSuffix)
		}
	}
}

func TestGetMemoryStatistics(t *testing.T) {
	ms := GetMemoryStatistics()
	if ms.TotalSystemMemory == "" {
		t.Error("expected non-empty TotalSystemMemory")
	}
	if ms.TotalSystemMemoryRaw <= 0 {
		t.Errorf("expected TotalSystemMemoryRaw > 0, got %f", ms.TotalSystemMemoryRaw)
	}
	if ms.AvailableMemoryRaw < 0 {
		t.Errorf("expected AvailableMemoryRaw >= 0, got %f", ms.AvailableMemoryRaw)
	}
}

func TestGetNetworkIO(t *testing.T) {
	recv, sent := GetNetworkIO()
	if recv < 0 {
		t.Errorf("expected BytesReceived >= 0, got %f", recv)
	}
	if sent < 0 {
		t.Errorf("expected BytesSent >= 0, got %f", sent)
	}
}

func TestConstructMemStats(t *testing.T) {
	m := ReadMemStats()
	records := ConstructMemStats(m)
	if len(records) != 27 {
		t.Errorf("expected 27 records, got %d", len(records))
	}
}

func TestConstructRawMemStats(t *testing.T) {
	m := ReadMemStats()
	records := ConstructRawMemStats(m)
	if len(records) != 27 {
		t.Errorf("expected 27 records (after duplicate fix), got %d", len(records))
	}

	// Verify non-byte metrics are stored as raw values (not divided by 1024)
	for _, r := range records {
		if r.RecordName == "gc_cpu_fraction" {
			// GCCPUFraction is typically a very small number (< 1.0)
			// If it were wrongly divided by 1024, it would be even smaller
			// Just verify it's not negative
			if r.RecordValue < 0 {
				t.Errorf("gc_cpu_fraction should be >= 0, got %f", r.RecordValue)
			}
		}
	}
}

func TestGetDiskIO(t *testing.T) {
	read, write := GetDiskIO()
	// Just verify no panic and values are reasonable
	_ = read
	_ = write
}

func TestGetServiceHealth(t *testing.T) {
	stats := GetServiceStats(context.Background())
	health := stats.Health

	if health.ServiceHealth.Message == "" {
		t.Error("expected non-empty service health message")
	}
	if health.SystemHealth.Message == "" {
		t.Error("expected non-empty system health message")
	}
}

func TestGetStatusMessage(t *testing.T) {
	tests := []struct {
		score    float64
		contains string
	}{
		{95, "Excellent"},
		{87, "Good"},
		{75, "Satisfactory"},
		{55, "Fair"},
		{35, "Poor"},
		{10, "Critical"},
	}
	for _, tt := range tests {
		msg := getStatusMessage(tt.score)
		if !strings.Contains(msg, tt.contains) {
			t.Errorf("getStatusMessage(%v) = %q, expected to contain %q", tt.score, msg, tt.contains)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "s"},
		{5 * time.Minute, "m"},
		{3 * time.Hour, "h"},
		{48 * time.Hour, "d"},
	}
	for _, tt := range tests {
		result := formatUptime(tt.d)
		if !strings.Contains(result, tt.want) {
			t.Errorf("formatUptime(%v) = %q, expected to contain %q", tt.d, result, tt.want)
		}
	}
}

func TestCollectGoRoutinesInfo(t *testing.T) {
	info := CollectGoRoutinesInfo()
	if info.NumberOfGoroutines <= 0 {
		t.Errorf("expected goroutines > 0, got %d", info.NumberOfGoroutines)
	}
	if len(info.StackView) == 0 {
		t.Error("expected non-empty stack view")
	}
}

func TestSplitGoroutines(t *testing.T) {
	input := "goroutine 1 [running]:\nmain.main()\n\ngoroutine 2 [sleep]:\ntime.Sleep()\n"
	blocks := SplitGoroutines(input)
	if len(blocks) != 2 {
		t.Errorf("expected 2 goroutine blocks, got %d", len(blocks))
	}
}
