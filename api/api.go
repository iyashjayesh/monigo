package api

import (
	"encoding/json"
	"fmt"
	"log"
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

// GetServiceInfoAPI returns the service metrics detailed information
func NewCoreStatistics(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()
	if fieldDescription == nil {
		log.Println("Field Description is nil")
		fieldDescription = common.ConstructJsonFieldDescription()
	}

	serviceStats := core.GetNewServiceStats()

	jsonMetrics, _ := json.Marshal(serviceStats)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMetrics))
	log.Println("Time taken to get the service stats Final: ", time.Since(startTime))
}

func GetFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "MB" // Default unit
	}

	// Convert bytes to different units
	bytesToUnit := func(bytes uint64) float64 {
		switch unit {
		case "KB":
			return float64(bytes) / 1024.0
		case "MB":
			return float64(bytes) / 1048576.0
		default: // "bytes"
			return float64(bytes)
		}
	}

	functionsMetrics := core.GetLocalFunctionMetrics()

	var results string
	mu.Lock()
	for name, metrics := range functionsMetrics {
		results += fmt.Sprintf(
			"Function: %s\nFunction Ran At: %s\nCPU Profile: %s\nExecution Time: %s\nMemory Usage: %.2f %s\nGoroutines: %d\n\n",
			name,
			metrics.FunctionLastRanAt.Format(time.RFC3339),
			metrics.CPUProfile,
			metrics.ExecutionTime,
			bytesToUnit(metrics.MemoryUsage),
			unit,
			metrics.GoroutineCount,
		)
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(results))
}

func GetGoRoutinesStats(w http.ResponseWriter, r *http.Request) {
	jsonGoRoutinesStats, _ := json.Marshal(core.CollectGoRoutinesInfo())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonGoRoutinesStats))
}

// func GetServiceMetricsFromStorage(w http.ResponseWriter, r *http.Request) {

// 	var req models.FetchDataPoints

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Failed to decode request", http.StatusBadRequest)
// 		return
// 	}

// 	startTime, err := time.Parse(time.RFC3339, req.StartTime)
// 	if err != nil {
// 		http.Error(w, "Invalid start time", http.StatusBadRequest)
// 		return
// 	}

// 	endTime, err := time.Parse(time.RFC3339, req.EndTime)
// 	if err != nil {
// 		http.Error(w, "Invalid end time", http.StatusBadRequest)
// 		return
// 	}

// 	labelName := "host"
// 	labelValue := "server1"

// 	var datapointRes []models.DataPointsInfo
// 	for _, fieldName := range req.FieldName {
// 		datapoints, err := timeseries.GetDataPoints(fieldName, []tstorage.Label{{Name: labelName, Value: labelValue}}, startTime.Unix(), endTime.Unix())
// 		if err != nil {
// 			http.Error(w, "Failed to get data points", http.StatusInternalServerError)
// 			return
// 		}

// 		datapointRes = append(datapointRes, models.DataPointsInfo{
// 			FieldName: fieldName,
// 			Data:      datapoints,
// 		})
// 	}

// 	// above "datapointRes" has all the data points for the requested fields and time range

// 	// 	{
// 	//     "time": "2024-09-06T09:07:33.770Z",
// 	//     "value": {
// 	//         "HeapAlloc": 6.47,
// 	//         "HeapSys": 13.37,
// 	//         "HeapInuse": 9.99,
// 	//         "HeapIdle": 4.77,
// 	//         "HeapReleased": 2.63
// 	//     }
// 	// }
// 	// but I want soemthign like above
// 	// how to acheive?

// 	jsonDP, err := json.Marshal(datapointRes)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(jsonDP)
// }

// func GetServiceMetricsFromStorage(w http.ResponseWriter, r *http.Request) {
// 	var req models.FetchDataPoints

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Failed to decode request", http.StatusBadRequest)
// 		return
// 	}

// 	startTime, err := time.Parse(time.RFC3339, req.StartTime)
// 	if err != nil {
// 		http.Error(w, "Invalid start time", http.StatusBadRequest)
// 		return
// 	}

// 	endTime, err := time.Parse(time.RFC3339, req.EndTime)
// 	if err != nil {
// 		http.Error(w, "Invalid end time", http.StatusBadRequest)
// 		return
// 	}

// 	labelName := "host"
// 	labelValue := "server1"

// 	// Map to group data points by timestamp
// 	dataByTimestamp := make(map[int64]map[string]float64)

// 	for _, fieldName := range req.FieldName {
// 		datapoints, err := timeseries.GetDataPoints(fieldName, []tstorage.Label{{Name: labelName, Value: labelValue}}, startTime.Unix(), endTime.Unix())
// 		if err != nil {
// 			http.Error(w, "Failed to get data points", http.StatusInternalServerError)
// 			return
// 		}

// 		for _, dp := range datapoints {
// 			if _, exists := dataByTimestamp[dp.Timestamp]; !exists {
// 				dataByTimestamp[dp.Timestamp] = make(map[string]float64)
// 			}
// 			dataByTimestamp[dp.Timestamp][fieldName] = dp.Value
// 		}
// 	}

// 	// Prepare the final response structure
// 	var result []map[string]interface{}
// 	for timestamp, values := range dataByTimestamp {
// 		result = append(result, map[string]interface{}{
// 			"time":  time.Unix(timestamp, 0).UTC().Format(time.RFC3339),
// 			"value": values,
// 		})
// 	}

// 	log.Println("Data points fetched successfully, " + labelName + ": " + labelValue)

// 	jsonDP, err := json.Marshal(result)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(jsonDP)
// }

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
		log.Println("Start time is before service start time, setting start time to service start time")
		log.Println("Service Start Time: " + serviceStartTime.String() + " Requested Start Time: " + startTime.String())
		startTime = serviceStartTime
	}

	log.Println("\n")
	log.Println("Request Fields: ", req.FieldName)
	log.Println("Start Time: " + startTime.String() + " End Time: " + endTime.String())
	log.Println("Start Time Unix: ", startTime.Unix(), " End Time Unix: ", endTime.Unix())
	log.Println("\n")

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
			// AcctualfieldName := NameMap[fieldName]

			if _, ok := NameMap[fieldName]; ok {
				// log.Println("Field Name Found in Map")
				dataByTimestamp[dp.Timestamp][NameMap[fieldName]] = dp.Value
			} else {
				// log.Println("Field Name Not Found in Map")
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

	log.Println("Data points fetched successfully, " + labelName + ": " + labelValue)

	jsonDP, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDP)
}

// func GetMetricsInfo(w http.ResponseWriter, r *http.Request) {

// 	// /get-metrics?fields=service-info

// 	fields := r.URL.Query().Get("fields")
// 	if fields == "" {
// 		http.Error(w, "Fields parameter is required", http.StatusBadRequest)
// 		return
// 	}

// 	switch fields {
// 	case "service-info":
// 		GetServiceInfoAPI(w, r)
// 	case "service-stats":
// 		GetMetrics(w, r)
// 	default:
// 		http.Error(w, "Invalid fields parameter", http.StatusBadRequest)
// 	}

// }

// func ProfileHandler(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("Generating profile\n")
// 	name := r.URL.Query().Get("name")
// 	if name == "" {
// 		http.Error(w, "Name parameter is required", http.StatusBadRequest)
// 		return
// 	}

// 	profilesFolderPath := fmt.Sprintf("%s/profiles", common.GetBasePath())

// 	cmd := exec.Command("go", "tool", "pprof", "-svg", profilesFolderPath)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		errMsg := fmt.Sprintf("failed to generate profile, given path %s, error: %v", profilesFolderPath, err)
// 		http.Error(w, errMsg, http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "image/svg+xml")
// 	if _, err := w.Write(output); err != nil {
// 		http.Error(w, "Failed to write response", http.StatusInternalServerError)
// 	}
// }
