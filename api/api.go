package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/models"
	"github.com/iyashjayesh/monigo/timeseries"
)

var (
	fieldDescription = map[string]string{}
	fieldDesOnce     = sync.Once{}
)

func init() {
	fieldDesOnce.Do(func() {
		fieldDescription = common.ConstructJsonFieldDescription()
	}) // This will be called only once
}

// GetServiceInfoAPI returns the service information
func GetServiceInfoAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(common.GetServiceInfo()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetServiceStatistics returns the service metrics detailed information
func GetServiceStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(core.GetServiceStats(r.Context())); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetGoRoutinesStats returns the goroutine statistics
func GetGoRoutinesStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(core.CollectGoRoutinesInfo()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

var NameMap = map[string]string{
	"heap_alloc":      "HeapAlloc",
	"heap_sys":        "HeapSys",
	"heap_inuse":      "HeapInuse",
	"heap_idle":       "HeapIdle",
	"heap_released":   "HeapReleased",
	"stack_inuse":     "StackInuse",
	"stack_sys":       "StackSys",
	"pause_total_ns":  "PauseTotalNs",
	"num_gc":          "NumGC",
	"gc_cpu_fraction": "GCCPUFraction",
	"m_span_inuse":    "MSpanInuse",
	"m_span_sys":      "MSpanSys",
	"m_cache_inuse":   "MCacheInuse",
	"m_cache_sys":     "MCacheSys",
	"buck_hash_sys":   "BuckHashSys",
	"gc_sys":          "GCSys",
	"other_sys":       "OtherSys",
}

// GetServiceMetricsFromStorage returns the service metrics from the storage
func GetServiceMetricsFromStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.FetchDataPoints
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	serviceStartTime := common.GetServiceStartTime()

	if startTime.Before(serviceStartTime) {
		startTime = serviceStartTime
	}

	hostLabel := timeseries.GetHostLabel()

	dataByTimestamp := make(map[int64]map[string]float64)

	for _, fieldName := range req.FieldName {
		datapoints, err := timeseries.GetDataPoints(fieldName, []timeseries.Label{hostLabel}, startTime.Unix(), endTime.Unix())
		if err != nil {
			http.Error(w, "Failed to get data points", http.StatusInternalServerError)
			return
		}

		for _, dp := range datapoints {
			if _, exists := dataByTimestamp[dp.Timestamp]; !exists {
				dataByTimestamp[dp.Timestamp] = make(map[string]float64)
			}
			if _, ok := NameMap[fieldName]; ok {
				dataByTimestamp[dp.Timestamp][NameMap[fieldName]] = dp.Value
			} else {
				dataByTimestamp[dp.Timestamp][fieldName] = dp.Value
			}
		}
	}

	var result []map[string]interface{}
	for timestamp, values := range dataByTimestamp {
		result = append(result, map[string]interface{}{
			"time":  time.Unix(timestamp, 0).UTC().Format(time.RFC3339Nano),
			"value": values,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["time"].(string) < result[j]["time"].(string)
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode data points", http.StatusInternalServerError)
	}
}

// GetReportData returns the report data
func GetReportData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqObj models.ReportsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqObj); err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, reqObj.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, reqObj.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}

	serviceStartTime := common.GetServiceStartTime()

	if startTime.Before(serviceStartTime) {
		startTime = serviceStartTime
	}

	var fieldNameList []string
	switch reqObj.Topic {
	case "LoadStatistics":
		fieldNameList = []string{"overall_load_of_service", "service_cpu_load", "service_memory_load", "system_cpu_load", "system_memory_load"}
	case "CPUStatistics":
		fieldNameList = []string{"total_cores", "cores_used_by_service", "cores_used_by_system"}
	case "MemoryStatistics":
		fieldNameList = []string{"total_system_memory", "memory_used_by_system", "memory_used_by_service", "available_memory", "gc_pause_duration", "stack_memory_usage"}
	case "MemoryProfile":
		fieldNameList = []string{"heap_alloc_by_service", "heap_alloc_by_system", "total_alloc_by_service", "total_memory_by_os"}
	case "NetworkIO":
		fieldNameList = []string{"bytes_sent", "bytes_received"}
	case "OverallHealth":
		fieldNameList = []string{"service_health_percent", "system_health_percent"}
	default:
		http.Error(w, "Unknown topic", http.StatusBadRequest)
		return
	}

	hostLabel := timeseries.GetHostLabel()

	dataByTimestamp := make(map[int64]map[string]float64)
	for _, fieldName := range fieldNameList {
		datapoints, err := timeseries.GetDataPoints(fieldName, []timeseries.Label{hostLabel}, startTime.Unix(), endTime.Unix())
		if err != nil {
			http.Error(w, "Failed to get data points", http.StatusInternalServerError)
			return
		}

		for _, dp := range datapoints {
			if _, exists := dataByTimestamp[dp.Timestamp]; !exists {
				dataByTimestamp[dp.Timestamp] = make(map[string]float64)
			}
			dataByTimestamp[dp.Timestamp][fieldName] = dp.Value
		}

	}

	var result []map[string]interface{}
	for timestamp, values := range dataByTimestamp {
		result = append(result, map[string]interface{}{
			"time":  time.Unix(timestamp, 0).UTC().Format(time.RFC3339Nano),
			"value": values,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i]["time"].(string) < result[j]["time"].(string)
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode report data", http.StatusInternalServerError)
	}
}

// GetFunctionTraceDetails returns the function trace details
func GetFunctionTraceDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(core.FunctionTraceDetails()); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// ViewFunctionMetrics returns detailed function metrics for a specific function
// GET /monigo/api/v1/function-details?name=FunctionName&reportType=text
func ViewFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	reportType := r.URL.Query().Get("reportType")

	if name == "" {
		http.Error(w, "Function name is required to get metrics", http.StatusBadRequest)
		return
	}

	if reportType == "" {
		reportType = "text"
	}

	metrics := core.FunctionTraceDetails()[name]
	if metrics == nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(core.ViewFunctionMetrics(name, reportType, metrics)); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
