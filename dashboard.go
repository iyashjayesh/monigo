// Functions for serving the dashboard and handling HTTP requests.
package monigo

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

var (
	//go:embed static/*
	staticFiles      embed.FS
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	Db               *bolt.DB
	basePath         string
	serviceInfo      ServiceInfo
	dbObj            *DBWrapper
)

func init() {
	Once.Do(func() {
		log.Println("Initializing the DB")
		basePath = GetBasePath()
		var err error
		dbObj, err = connectDb()
		if err != nil {
			log.Fatalf("Error connecting to database: %v\n", err)
		}
	})
}

func Start(addr int, serviceName string) {
	// Store the service info
	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVerison = runtime.Version()
	serviceInfo.TimeStamp = serviceStartTime

	dbObj.StoreServiceInfo(&serviceInfo)
	serviceInfo, err := dbObj.GetServiceInfo(serviceInfo.ServiceName)
	if err != nil {
		log.Fatalf("Error getting service info: %v\n", err)
	}
	log.Printf("Service Name: %s\nService Start Time: %s\nGo Version: %s\nTime Stamp: %s\n",
		serviceInfo.ServiceName, serviceInfo.ServiceStartTime, serviceInfo.GoVerison, serviceInfo.TimeStamp)

	go StartDashboard(addr)
}

func StartDashboard(addr int) {

	log.Println("Starting the dashboard")

	http.HandleFunc("/", serveHtmlSite)
	http.HandleFunc("/metrics", getMetrics)
	http.HandleFunc("/function-metrics", getFunctionMetrics)
	http.HandleFunc("/generate-function-metrics", profileHandler)

	// API to fetch the service metrics
	http.HandleFunc("/get-service-info", GetServiceInfoAPI)
	// http.HandleFunc("/get-service-metrics", GetServiceMetricsFromMonigoDbData)
	// http.HandleFunc("/get-function-info", getFunctionMetricsAPI)

	fmt.Printf("Starting dashboard at http://localhost:%d\n", addr)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", addr), nil); err != nil {
		log.Fatalf("Error starting the dashboard: %v\n", err)
	}
}

func serveHtmlSite(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		file, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Could not load index.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(file)
		return
	}
	http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))).ServeHTTP(w, r)
}

func getMetrics(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "MB" // Default Unit
	}

	requestCount, totalDuration, memStats := GetServiceMetricsFromMonigoDb()
	serviceStat := GetProcessSats()

	// var serviceMetrics ServiceMetrics

	// serviceMetrics.Id = uuid.New()
	// serviceMetrics.Load = fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent)
	// serviceMetrics.Cores = fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores) + "PC / " +
	// 	fmt.Sprintf("%.2f", serviceStat.SystemUsedCores) + "SC / " +
	// 	strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
	// 	strconv.Itoa(serviceStat.TotalCores) + "C"
	// serviceMetrics.MemoryUsed = fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)
	// serviceMetrics.UpTime = time.Since(serviceInfo.ServiceStartTime)
	// serviceMetrics.NumberOfReqServerd = requestCount
	// serviceMetrics.TotalDurationTookByAPI = totalDuration
	// serviceMetrics.TimeStamp = time.Now()

	// dbObj.StoreServiceRuntimeMetrics(&serviceMetrics)

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

	metrics := fmt.Sprintf(
		"Service Name: %s\nService Start Time: %s\nGoroutines: %d\nRequests: %d\nTotal Duration: %s\n\nMemory Usage (%s):\nAlloc: %.2f %s\nTotalAlloc: %.2f %s\nSys: %.2f %s\nHeapAlloc: %.2f %s\nHeapSys: %.2f %s\nGo Version: %s\n Load: %s\nCores: %s\n Memory Used: %s\n",
		serviceInfo.ServiceName,
		serviceStartTime.Format(time.RFC3339),
		GetGoroutineCount(),
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

func getFunctionMetrics(w http.ResponseWriter, r *http.Request) {
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

	var results string
	mu.Lock()
	for name, metrics := range functionMetrics {
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

func profileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Generating profile\n")
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	profilesFolderPath := fmt.Sprintf("%s/profiles", basePath)

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

// //
func GetServiceInfoAPI(w http.ResponseWriter, r *http.Request) {
	serviceInfo, err := dbObj.GetServiceInfo(serviceInfo.ServiceName)
	if err != nil {
		log.Println("Error getting service info:", err)
	}

	jsonServiceInfo, err := json.Marshal(serviceInfo)
	if err != nil {
		log.Println("Error marshalling service info:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonServiceInfo)
}

func GetServiceMetricsFromMonigoDbData() *ServiceMetrics {

	requestCount, totalDuration, memStats := GetServiceMetricsFromMonigoDb()
	serviceStat := GetProcessSats()

	var serviceMetrics ServiceMetrics

	serviceMetrics.Id = uuid.New()
	serviceMetrics.Load = fmt.Sprintf("%.2f", serviceStat.ProcCPUPercent)
	serviceMetrics.Cores = fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores) + "PC / " +
		fmt.Sprintf("%.2f", serviceStat.SystemUsedCores) + "SC / " +
		strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
		strconv.Itoa(serviceStat.TotalCores) + "C"
	serviceMetrics.MemoryUsed = fmt.Sprintf("%.2f", serviceStat.ProcMemPercent)
	serviceMetrics.UpTime = time.Since(serviceInfo.ServiceStartTime)
	serviceMetrics.NumberOfReqServerd = requestCount
	serviceMetrics.TotalDurationTookByAPI = totalDuration
	serviceMetrics.GoRoutines = GetGoroutineCount()
	serviceMetrics.TotalAlloc = memStats.TotalAlloc
	serviceMetrics.MemoryAllocSys = memStats.Sys
	serviceMetrics.HeapAlloc = memStats.HeapAlloc
	serviceMetrics.HeapAllocSys = memStats.HeapSys
	serviceMetrics.TimeStamp = time.Now()

	return &serviceMetrics
}

func ShowRuntimeMetrics() {
	// Store the Service Metrics
	metricsInfo, err := dbObj.GetServiceMetricsFromMonigoDb()
	if err != nil {
		log.Fatalf("Error getting service metrics: %v\n", err)
	}

	log.Printf("Service Metrics: %+v\n", metricsInfo)
}
