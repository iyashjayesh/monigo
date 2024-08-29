package api

import (
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
)

var (
	mu sync.Mutex = sync.Mutex{}
)

func GetServiceInfoAPI(w http.ResponseWriter, r *http.Request) {

	// dbObj := monigodb.GetDbInstance()
	// serviceInfo := dbObj.GetServiceDetails()

	// serviceInfo, err := dbObj.GetServiceInfo(serviceInfo.ServiceName)
	// if err != nil {
	// 	log.Println("Error getting service info:", err)
	// }

	// jsonServiceInfo, err := json.Marshal(serviceInfo)
	// if err != nil {
	// 	log.Println("Error marshalling service info:", err)
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(jsonServiceInfo)
}


func GetMetrics(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "MB" // Default Unit
	}

	requestCount, totalDuration, memStats := core.GetServiceMetrics()
	serviceStat := core.GetProcessSats()

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

	metrics := fmt.Sprintf(
		"Service Name: %s\nService Start Time: %s\nGoroutines: %d\nRequests: %d\nTotal Duration: %s\n\nMemory Usage (%s):\nAlloc: %.2f %s\nTotalAlloc: %.2f %s\nSys: %.2f %s\nHeapAlloc: %.2f %s\nHeapSys: %.2f %s\nGo Version: %s\n Load: %s\nCores: %s\n Memory Used: %s\n",
		// serviceInfo.ServiceName,
		// serviceStartTime.Format(time.RFC3339),
		"name",
		"serviceStartTime",
		runtimeGoRoutine,
		requestCount,
		totalDuration,
		unit,
		bytesToUnit(memStats.Alloc),
		unit,
		bytesToUnit(memStats.TotalAlloc),
		unit,
		bytesToUnit(memStats.Sys),
		unit,
		bytesToUnit(memStats.HeapAlloc),
		unit,
		bytesToUnit(memStats.HeapSys),
		unit,
		runtime.Version(),
		fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent),
		core,
		memoryUsed,
	)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
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
