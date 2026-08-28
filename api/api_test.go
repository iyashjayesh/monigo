package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/models"
)

func init() {
	common.SetServiceInfo("test-service", time.Now(), runtime.Version(), 1234, "7d")
	core.ConfigureServiceThresholds(&models.ServiceHealthThresholds{
		MaxCPUUsage:    95,
		MaxMemoryUsage: 95,
		MaxGoRoutines:  1000,
	})
}

func TestGetServiceInfoAPI(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/service-info", nil)
	w := httptest.NewRecorder()
	GetServiceInfoAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var info models.ServiceInfo
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if info.ServiceName != "test-service" {
		t.Errorf("expected service name 'test-service', got %q", info.ServiceName)
	}
}

func TestGetServiceInfoAPI_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/service-info", nil)
	w := httptest.NewRecorder()
	GetServiceInfoAPI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetServiceStatistics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	GetServiceStatistics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var stats models.ServiceStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if stats.CoreStatistics.Goroutines <= 0 {
		t.Error("expected goroutines > 0")
	}
}

func TestGetServiceStatistics_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	GetServiceStatistics(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetGoRoutinesStats(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/go-routines-stats", nil)
	w := httptest.NewRecorder()
	GetGoRoutinesStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var stats models.GoRoutinesStatistic
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if stats.NumberOfGoroutines <= 0 {
		t.Error("expected goroutines > 0")
	}
}

func TestGetFunctionTraceDetails(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/function", nil)
	w := httptest.NewRecorder()
	GetFunctionTraceDetails(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetFunctionTraceDetails_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/function", nil)
	w := httptest.NewRecorder()
	GetFunctionTraceDetails(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestViewFunctionMetrics_MissingName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/function-details", nil)
	w := httptest.NewRecorder()
	ViewFunctionMetrics(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestViewFunctionMetrics_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/function-details?name=nonexistent", nil)
	w := httptest.NewRecorder()
	ViewFunctionMetrics(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetServiceMetricsFromStorage_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/service-metrics", nil)
	w := httptest.NewRecorder()
	GetServiceMetricsFromStorage(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetServiceMetricsFromStorage_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/service-metrics", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	GetServiceMetricsFromStorage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetReportData_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/reports", nil)
	w := httptest.NewRecorder()
	GetReportData(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGetReportData_UnknownTopic(t *testing.T) {
	body := `{"topic":"UnknownTopic","start_time":"2026-01-01T00:00:00Z","end_time":"2026-01-02T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/reports", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	GetReportData(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown topic, got %d", w.Code)
	}
}

func TestGetReportData_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/monigo/api/v1/reports", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()
	GetReportData(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// The dashboard charts must plot unformatted byte counts, not the display
// strings. The string fields carry a unit suffix ("6.82 MB", "15.2 GB"), so
// parsing a number out of one drops the unit and puts values of different
// magnitudes on the same axis -- which is how the Memory Distribution pie came
// to render a service using megabytes as a comparable slice to a system using
// gigabytes.
func TestServiceMetricsExposesRawMemoryBytes(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	GetServiceStatistics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	memStats, ok := payload["memory_statistics"].(map[string]interface{})
	if !ok {
		t.Fatal("response has no memory_statistics object")
	}

	// Fields the charts depend on. Absence means a chart silently falls back to
	// zero or to a formatted string.
	required := []string{
		"total_system_memory_bytes",
		"memory_used_by_system_bytes",
		"memory_used_by_service_bytes",
		"available_memory_bytes",
		"stack_memory_usage_bytes",
		"gc_pause_duration_ms",
	}
	for _, field := range required {
		v, present := memStats[field]
		if !present {
			t.Errorf("memory_statistics.%s is missing; charts cannot plot without it", field)
			continue
		}
		if _, isNumber := v.(float64); !isNumber {
			t.Errorf("memory_statistics.%s = %T, want a number (a string would carry a unit suffix)", field, v)
		}
	}

	// Total system memory is the one value that is always non-zero on a working
	// host, so it is the sanity check that these are real readings.
	if total, _ := memStats["total_system_memory_bytes"].(float64); total <= 0 {
		t.Errorf("total_system_memory_bytes = %v, want > 0", total)
	}

	// Byte counts must be plottable against each other: the service cannot use
	// more than the system total.
	svc, _ := memStats["memory_used_by_service_bytes"].(float64)
	total, _ := memStats["total_system_memory_bytes"].(float64)
	if total > 0 && svc > total {
		t.Errorf("memory_used_by_service_bytes (%v) exceeds total_system_memory_bytes (%v); "+
			"these are not in the same unit", svc, total)
	}
}

// raw_mem_stats_records is what the Heap Memory Usage chart reads. Unlike
// mem_stats_records -- whose record_value is a display number with its unit in a
// separate record_unit field -- every entry here is in one consistent unit.
func TestServiceMetricsExposesRawHeapRecords(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/monigo/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	GetServiceStatistics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	memStats, _ := payload["memory_statistics"].(map[string]interface{})
	records, ok := memStats["raw_mem_stats_records"].([]interface{})
	if !ok {
		t.Fatal("memory_statistics.raw_mem_stats_records is missing or not an array")
	}

	byName := map[string]float64{}
	for _, r := range records {
		rec, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rec["record_name"].(string)
		value, isNumber := rec["record_value"].(float64)
		if !isNumber {
			t.Errorf("raw_mem_stats_records[%q].record_value = %T, want a number", name, rec["record_value"])
			continue
		}
		byName[name] = value
	}

	// The five series the heap chart plots.
	for _, key := range []string{"heap_alloc", "heap_sys", "heap_idle", "heap_inuse", "heap_released"} {
		if _, present := byName[key]; !present {
			t.Errorf("raw_mem_stats_records is missing %q, which the heap chart plots", key)
		}
	}

	// heap_sys is everything the heap obtained from the OS, so it cannot be
	// smaller than the part currently allocated. If it is, the two are not in
	// the same unit -- exactly the bug this replaced.
	if alloc, okA := byName["heap_alloc"]; okA {
		if sys, okS := byName["heap_sys"]; okS && sys > 0 && alloc > sys {
			t.Errorf("heap_alloc (%v) exceeds heap_sys (%v); these are not in the same unit", alloc, sys)
		}
	}
}
