package monigo

import (
	"testing"
)

func TestBuilderValidBuild(t *testing.T) {
	m := NewBuilder().
		WithServiceName("test-service").
		WithPort(9090).
		WithStorageType("memory").
		WithSamplingRate(50).
		Build()

	if m.ServiceName != "test-service" {
		t.Errorf("expected 'test-service', got %q", m.ServiceName)
	}
	if m.DashboardPort != 9090 {
		t.Errorf("expected port 9090, got %d", m.DashboardPort)
	}
	if m.StorageType != "memory" {
		t.Errorf("expected 'memory', got %q", m.StorageType)
	}
	if m.SamplingRate != 50 {
		t.Errorf("expected sampling rate 50, got %d", m.SamplingRate)
	}
}

func TestBuilderMissingServiceName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing ServiceName")
		}
	}()

	NewBuilder().WithPort(8080).Build()
}

func TestBuilderInvalidPort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid port")
		}
	}()

	NewBuilder().WithServiceName("test").WithPort(-1).Build()
}

func TestBuilderInvalidStorageType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid storage type")
		}
	}()

	NewBuilder().WithServiceName("test").WithStorageType("redis").Build()
}

func TestBuilderDefaultStorageType(t *testing.T) {
	// Empty storage type should be allowed (defaults at runtime)
	m := NewBuilder().WithServiceName("test").Build()
	if m.StorageType != "" {
		t.Errorf("expected empty storage type, got %q", m.StorageType)
	}
}

func TestBuilderAllOptions(t *testing.T) {
	m := NewBuilder().
		WithServiceName("full-test").
		WithPort(3000).
		WithRetentionPeriod("30d").
		WithDataPointsSyncFrequency("1m").
		WithTimeZone("UTC").
		WithMaxCPUUsage(80).
		WithMaxMemoryUsage(80).
		WithMaxGoRoutines(500).
		WithSamplingRate(10).
		WithStorageType("disk").
		WithHeadless(true).
		WithCustomBaseAPIPath("/custom/api").
		Build()

	if m.DataRetentionPeriod != "30d" {
		t.Errorf("expected '30d', got %q", m.DataRetentionPeriod)
	}
	if m.MaxCPUUsage != 80 {
		t.Errorf("expected 80, got %f", m.MaxCPUUsage)
	}
	if !m.Headless {
		t.Error("expected headless true")
	}
	if m.CustomBaseAPIPath != "/custom/api" {
		t.Errorf("expected '/custom/api', got %q", m.CustomBaseAPIPath)
	}
}
