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

func TestBuilderThresholdBounds(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*MonigoBuilder) *MonigoBuilder
		wantPanic bool
	}{
		{"negative CPU", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxCPUUsage(-1) }, true},
		{"CPU above 100", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxCPUUsage(101) }, true},
		{"CPU at 100", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxCPUUsage(100) }, false},
		{"negative memory", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxMemoryUsage(-0.5) }, true},
		{"memory above 100", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxMemoryUsage(150) }, true},
		{"memory at 0", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxMemoryUsage(0) }, false},
		{"negative goroutines", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxGoRoutines(-1) }, true},
		{"goroutines at 0", func(b *MonigoBuilder) *MonigoBuilder { return b.WithMaxGoRoutines(0) }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Error("expected Build() to panic")
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("expected Build() to succeed, panicked with: %v", r)
				}
			}()

			tt.configure(NewBuilder().WithServiceName("test")).Build()
		})
	}
}

func TestBuilderAlertWebhookValidation(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantPanic bool
	}{
		{"https", "https://hooks.example.com/services/abc", false},
		{"http", "http://localhost:9000/alerts", false},
		{"unset", "", false},
		{"no scheme", "hooks.example.com/services/abc", true},
		{"wrong scheme", "ftp://hooks.example.com/abc", true},
		{"shell-ish", "; rm -rf /", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Errorf("expected Build() to panic for %q", tt.url)
				}
				if !tt.wantPanic && r != nil {
					t.Errorf("expected Build() to accept %q, panicked with: %v", tt.url, r)
				}
			}()

			m := NewBuilder().WithServiceName("test").WithAlertWebhook(tt.url).Build()
			if m.AlertWebhookURL != tt.url {
				t.Errorf("AlertWebhookURL = %q, want %q", m.AlertWebhookURL, tt.url)
			}
		})
	}
}
