package common

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBytesToUnit(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0.00 B"},
		{512, "512.00 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}
	for _, tt := range tests {
		got := BytesToUnit(tt.input)
		if got != tt.want {
			t.Errorf("BytesToUnit(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConvertBytesToUnit(t *testing.T) {
	tests := []struct {
		bytes float64
		unit  string
		want  float64
	}{
		{1024, "KB", 1.0},
		{1048576, "MB", 1.0},
		{1073741824, "GB", 1.0},
	}
	for _, tt := range tests {
		got := ConvertBytesToUnit(tt.bytes, tt.unit)
		if got != tt.want {
			t.Errorf("ConvertBytesToUnit(%f, %q) = %f, want %f", tt.bytes, tt.unit, got, tt.want)
		}
	}
}

func TestConvertToMB(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1.00GB", 1024.0},
		{"1.00MB", 1.0},
		{"1024.00KB", 1.0},
		{"1.00TB", 1048576.0},
	}
	for _, tt := range tests {
		got, err := ConvertToMB(tt.input)
		if err != nil {
			t.Errorf("ConvertToMB(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ConvertToMB(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	if DefaultIfEmpty("", "default") != "default" {
		t.Error("expected default for empty string")
	}
	if DefaultIfEmpty("value", "default") != "value" {
		t.Error("expected value for non-empty string")
	}
}

func TestDefaultFloatIfZero(t *testing.T) {
	if DefaultFloatIfZero(0, 5.5) != 5.5 {
		t.Error("expected 5.5 for zero input")
	}
	if DefaultFloatIfZero(3.3, 5.5) != 3.3 {
		t.Error("expected 3.3 for non-zero input")
	}
}

func TestDefaultIntIfZero(t *testing.T) {
	if DefaultIntIfZero(0, 10) != 10 {
		t.Error("expected 10 for zero input")
	}
	if DefaultIntIfZero(7, 10) != 7 {
		t.Error("expected 7 for non-zero input")
	}
}

func TestRoundFloat64(t *testing.T) {
	tests := []struct {
		value     float64
		precision int
		want      float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 4, 3.1416},
		{0, 2, 0},
	}
	for _, tt := range tests {
		got := RoundFloat64(tt.value, tt.precision)
		if got != tt.want {
			t.Errorf("RoundFloat64(%f, %d) = %f, want %f", tt.value, tt.precision, got, tt.want)
		}
	}
}

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cache.dat")

	// Save
	cache := Cache{Data: map[string]time.Time{
		"test-service": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	if err := cache.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}

	// Load
	loaded := Cache{Data: make(map[string]time.Time)}
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	startTime, exists := loaded.Data["test-service"]
	if !exists {
		t.Fatal("expected 'test-service' in loaded cache")
	}
	if !startTime.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected time: %v", startTime)
	}
}

func TestCacheLoadFromEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.dat")
	os.WriteFile(path, []byte{}, 0644)

	cache := Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(path); err != nil {
		t.Errorf("LoadFromFile on empty file should not error, got: %v", err)
	}
}

func TestGetBasePath(t *testing.T) {
	path := GetBasePath()
	if path == "" {
		t.Error("expected non-empty base path")
	}
}

func TestGetDataRetentionPeriod(t *testing.T) {
	SetServiceInfo("test", time.Now(), "go1.24", 1, "7d")
	d := GetDataRetentionPeriod()
	if d != 7*24*time.Hour {
		t.Errorf("expected 7 days, got %v", d)
	}
}

func TestConvertToReadableUnit(t *testing.T) {
	result := ConvertToReadableUnit(uint64(1048576))
	if result != "1.00 MB" {
		t.Errorf("expected '1.00 MB', got %q", result)
	}
}

func TestParseFloat64ToString(t *testing.T) {
	result := ParseFloat64ToString(3.14)
	if result != "3.14" {
		t.Errorf("expected '3.14', got %q", result)
	}
}

func TestGetProcessId(t *testing.T) {
	pid := GetProcessId()
	if pid <= 0 {
		t.Errorf("expected positive PID, got %d", pid)
	}
}
