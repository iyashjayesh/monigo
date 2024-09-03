package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
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

	var serviceStats models.NewServiceStats

	timeNow := time.Now()
	serviceStats.CoreStatistics = core.GetCoreStatistics()
	log.Println("Time taken to get the core statistics: ", time.Since(timeNow))

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		timeNow := time.Now()
		serviceStats.LoadStatistics = core.GetLoadStatistics()
		log.Println("Time taken to get the load statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		timeNow := time.Now()
		serviceStats.MemoryStatistics = core.GetMemoryStatistics()
		log.Println("Time taken to get the memory statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		timeNow := time.Now()
		serviceStats.CPUStatistics = core.GetCPUStatistics()
		log.Println("Time taken to get the CPU statistics: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		timeNow := time.Now()
		memStats := core.ReadMemStats()
		serviceStats.HeapAllocByService = common.BytesToUnit(memStats.HeapAlloc)
		serviceStats.HeapAllocBySystem = common.BytesToUnit(memStats.HeapSys)
		serviceStats.TotalAllocByService = common.BytesToUnit(memStats.TotalAlloc)
		serviceStats.TotalMemoryByOS = common.BytesToUnit(memStats.Sys)
		log.Println("Time taken to get the memory stats: ", time.Since(timeNow))
	}()

	go func() {
		defer wg.Done()
		timeNow := time.Now()
		serviceStats.NetworkIO.BytesReceived, serviceStats.NetworkIO.BytesSent = core.GetNetworkIO()
		log.Println("Time taken to get the network stats: ", time.Since(timeNow))
	}()

	wg.Wait()

	serviceStats.OverallHealth = core.GetServiceHealth(&serviceStats.LoadStatistics)
	// serviceStats.DiskIO = core.GetDiskIO()                                             // TODO: Need to implement this function

	jsonMetrics, _ := json.Marshal(serviceStats)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMetrics))
	log.Println("Time taken to get the service stats Final: ", time.Since(startTime))
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "MB" // Default Unit
	}

	requestCount, totalDuration := core.GetServiceMetrics()
	serviceStat := core.GetProcessSats()

	memStats := core.ReadMemStats()

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

	SystemUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.SystemUsedCores)
	ProcessUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores)

	core := ProcessUsedCoresToString + "PC / " +
		SystemUsedCoresToString + "SC / " +
		strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
		strconv.Itoa(serviceStat.TotalCores) + "C"

	// ProcMemPercent
	memoryUsed := fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)
	runtimeGoRoutine := runtime.NumGoroutine()
	serviceInfo := common.GetServiceInfo()

	// 7.051466958s
	uptime := time.Since(serviceInfo.ServiceStartTime)
	uptimeStr := fmt.Sprintf("%.2f s", uptime.Seconds())

	if uptime.Seconds() > 60 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Minutes())
	} else if uptime.Hours() > 60 {
		uptimeStr = fmt.Sprintf("%.2f h", uptime.Hours())
	} else if uptime.Hours() > 24 {
		uptimeStr = fmt.Sprintf("%.2f d", uptime.Hours()/24)
	} else if uptime.Hours() > 30*24 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Hours()/(30*24))
	} else if uptime.Hours() > 12*30*24 {
		uptimeStr = fmt.Sprintf("%.2f y", uptime.Hours()/(12*30*24))
	}

	metrics := struct {
		Goroutines    int    `json:"goroutines"`
		Requests      int64  `json:"requests"`
		TotalDuration string `json:"total_duration"`
		MemoryUsage   string `json:"memory_usage"`
		Alloc         string `json:"alloc"`
		TotalAlloc    string `json:"total_alloc"`
		Sys           string `json:"sys"`
		HeapAlloc     string `json:"heap_alloc"`
		HeapSys       string `json:"heap_sys"`
		Load          string `json:"load"`
		Cores         string `json:"cores"`
		MemoryUsed    string `json:"memory_used"`
		Uptime        string `json:"uptime"`
	}{
		Goroutines:    runtimeGoRoutine,
		Requests:      requestCount,
		TotalDuration: totalDuration.String(),
		MemoryUsage:   unit,
		Alloc:         fmt.Sprintf("%.2f", bytesToUnit(memStats.Alloc)),
		TotalAlloc:    fmt.Sprintf("%.2f", bytesToUnit(memStats.TotalAlloc)),
		Sys:           fmt.Sprintf("%.2f", bytesToUnit(memStats.Sys)),
		HeapAlloc:     fmt.Sprintf("%.2f", bytesToUnit(memStats.HeapAlloc)),
		HeapSys:       fmt.Sprintf("%.2f", bytesToUnit(memStats.HeapSys)),
		Load:          fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent) + "%",
		Cores:         core,
		MemoryUsed:    memoryUsed + "%",
		Uptime:        uptimeStr,
	}

	jsonMetrics, _ := json.Marshal(metrics)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMetrics))
}

func GetCoreStats(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "MB" // Default Unit
	}

	requestCount, totalDuration := core.GetServiceMetrics()
	// serviceStat := core.GetProcessSats()

	// memStats := core.ReadMemStats()
	// memStatsRecord := core.ConstructMemStats(memStats)

	// Convert bytes to different units
	// bytesToUnit := func(bytes uint64) float64 {
	// 	switch unit {
	// 	case "KB":
	// 		return float64(bytes) / 1024.0
	// 	case "MB":
	// 		return float64(bytes) / 1048576.0
	// 	case "GB":
	// 		return float64(bytes) / 1073741824.0
	// 	default: // "bytes"
	// 		return float64(bytes)
	// 	}
	// }

	// SystemUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.SystemUsedCores)
	// ProcessUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores)

	// coreStr := ProcessUsedCoresToString + "PC / " +
	// 	SystemUsedCoresToString + "SC / " +
	// 	strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
	// 	strconv.Itoa(serviceStat.TotalCores) + "C"

	// ProcMemPercent
	// memoryUsed := fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)
	runtimeGoRoutine := runtime.NumGoroutine()
	serviceInfo := common.GetServiceInfo()

	// 7.051466958s
	uptime := time.Since(serviceInfo.ServiceStartTime)
	uptimeStr := fmt.Sprintf("%.2f s", uptime.Seconds())

	if uptime.Seconds() > 60 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Minutes())
	} else if uptime.Hours() > 60 {
		uptimeStr = fmt.Sprintf("%.2f h", uptime.Hours())
	} else if uptime.Hours() > 24 {
		uptimeStr = fmt.Sprintf("%.2f d", uptime.Hours()/24)
	} else if uptime.Hours() > 30*24 {
		uptimeStr = fmt.Sprintf("%.2f m", uptime.Hours()/(30*24))
	} else if uptime.Hours() > 12*30*24 {
		uptimeStr = fmt.Sprintf("%.2f y", uptime.Hours()/(12*30*24))
	}

	metrics := models.ServiceCoreStats{
		Goroutines:   runtimeGoRoutine,
		RequestCount: requestCount,
		// Load:                       fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent) + "%",
		Memory:                     core.GetSystemMemoryInfo(),
		Uptime:                     uptimeStr,
		TotalDurationTookbyRequest: totalDuration.Seconds(),
		CPU:                        core.GetSystemCPUInfo(),
	}

	jsonMetrics, _ := json.Marshal(metrics)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonMetrics))
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

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Generating profile\n")
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	profilesFolderPath := fmt.Sprintf("%s/profiles", common.GetBasePath())

	cmd := exec.Command("go", "tool", "pprof", "-svg", profilesFolderPath)
	output, err := cmd.Output()
	if err != nil {
		errMsg := fmt.Sprintf("failed to generate profile, given path %s, error: %v", profilesFolderPath, err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write(output); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

type ReqObj struct {
	FieldName string `json:"field_name"`
	StartTime string `json:"start_time"` // "2006-01-02T15:04:05Z07:00"
	EndTime   string `json:"end_time"`   // "2006-01-02T15:04:05Z07:00"
}

func GetServiceMetricsFromStorage(w http.ResponseWriter, r *http.Request) {

	var req ReqObj
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

	datapoints, err := timeseries.GetDataPoints(req.FieldName, []tstorage.Label{{Name: "host", Value: "server1"}}, startTime.Unix(), endTime.Unix())
	if err != nil {
		http.Error(w, "Failed to get data points", http.StatusInternalServerError)
		return
	}

	jsonDP, err := json.Marshal(datapoints)
	if err != nil {
		http.Error(w, "Failed to marshal data points", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonDP)
}

func GetMetricsInfo(w http.ResponseWriter, r *http.Request) {

	// /get-metrics?fields=service-info

	fields := r.URL.Query().Get("fields")
	if fields == "" {
		http.Error(w, "Fields parameter is required", http.StatusBadRequest)
		return
	}

	switch fields {
	case "service-info":
		GetServiceInfoAPI(w, r)
	case "service-stats":
		GetMetrics(w, r)
	default:
		http.Error(w, "Invalid fields parameter", http.StatusBadRequest)
	}

}

func GetGoRoutinesStats(w http.ResponseWriter, r *http.Request) {
	jsonGoRoutinesStats, _ := json.Marshal(core.CollectGoRoutinesInfo())
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonGoRoutinesStats))
}
