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
