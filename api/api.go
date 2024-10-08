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
	"github.com/nakabonne/tstorage"
)

var (
	mu               sync.Mutex = sync.Mutex{}
	fieldDescription            = map[string]string{}
	fieldDesOnce                = sync.Once{}
)

func init() {
	fieldDesOnce.Do(func() {
		fieldDescription = common.ConstructJsonFieldDescription()
	}) // This will be called only once
}

// GetServiceInfoAPI returns the service information
func GetServiceInfoAPI(w http.ResponseWriter, r *http.Request) {
	jsonObjStr, _ := json.Marshal(common.GetServiceInfo())
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonObjStr)
}

// GetServiceStatistics returns the service metrics detailed information
func GetServiceStatistics(w http.ResponseWriter, r *http.Request) {
	if fieldDescription == nil {
		fieldDescription = common.ConstructJsonFieldDescription()
	}

	jsonMetrics, _ := json.Marshal(core.GetServiceStats())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMetrics))
}

func GetGoRoutinesStats(w http.ResponseWriter, r *http.Request) {
	jsonGoRoutinesStats, _ := json.Marshal(core.CollectGoRoutinesInfo())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonGoRoutinesStats))
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

	labelName := "host"
	labelValue := "server1"

	dataByTimestamp := make(map[int64]map[string]float64)

	for _, fieldName := range req.FieldName {
		datapoints, err := timeseries.GetDataPoints(fieldName, []tstorage.Label{{Name: labelName, Value: labelValue}}, startTime.Unix(), endTime.Unix())
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

	// sort the timingdataByTimestamp in ascending order
	sort.Slice(result, func(i, j int) bool {
		return result[i]["time"].(string) < result[j]["time"].(string)
	})

	jsonDP, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDP)
}

// GetReportData returns the report data
func GetReportData(w http.ResponseWriter, r *http.Request) {

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
	if reqObj.Topic == "LoadStatistics" {
		fieldNameList = []string{"overall_load_of_service", "service_cpu_load", "service_memory_load", "system_cpu_load", "system_memory_load"}
	} else if reqObj.Topic == "CPUStatistics" {
		fieldNameList = []string{"total_cores", "cores_used_by_service", "cores_used_by_system"}
	} else if reqObj.Topic == "MemoryStatistics" {
		fieldNameList = []string{"total_system_memory", "memory_used_by_system", "memory_used_by_service", "available_memory", "gc_pause_duration", "stack_memory_usage"}
	} else if reqObj.Topic == "MemoryProfile" {
		fieldNameList = []string{"heap_alloc_by_service", "heap_alloc_by_system", "total_alloc_by_service", "total_memory_by_os"}
	} else if reqObj.Topic == "NetworkIO" {
		fieldNameList = []string{"bytes_sent", "bytes_received"}
	} else if reqObj.Topic == "OverallHealth" {
		fieldNameList = []string{"service_health_percent", "system_health_percent"}
	}

	labelName := "host"
	labelValue := "server1"

	dataByTimestamp := make(map[int64]map[string]float64)
	for _, fieldName := range fieldNameList {

		datapoints, err := timeseries.GetDataPoints(fieldName, []tstorage.Label{{Name: labelName, Value: labelValue}}, startTime.Unix(), endTime.Unix())
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

	jsonDP, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDP)
}

// GetFunctionTraceDetails returns the function trace details
func GetFunctionTraceDetails(w http.ResponseWriter, r *http.Request) {
	jsonObjStr, _ := json.Marshal(core.FunctionTraceDetails())
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonObjStr)
}

// /monigo/api/v1/function-details?name=FunctionName&reportType=text
func ViewFunctionMaetrtics(w http.ResponseWriter, r *http.Request) {

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

	w.Header().Set("Content-Type", "text/plain")
	jsonResp, _ := json.Marshal(core.ViewFunctionMetrics(name, reportType, metrics))
	w.Write(jsonResp)
}
