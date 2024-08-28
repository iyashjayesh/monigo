package monigo

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/api"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/models"
	"github.com/iyashjayesh/monigo/timeseries"
	"github.com/nakabonne/tstorage"
)

var (
	//go:embed static/*
	staticFiles      embed.FS
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	BasePath         string
	serviceInfo      models.ServiceInfo
	mu               sync.Mutex = sync.Mutex{}
)

func init() {
	BasePath = GetBasePath()
	// dbObj = monigodb.GetDbInstance()
}

func Start(addr int, serviceName string) {
	// Store the service info
	serviceInfo.ServiceName = serviceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVersion = runtime.Version()
	serviceInfo.TimeStamp = serviceStartTime

	log.Printf("Service Name: %s\nService Start Time: %s\nGo Version: %s\nTime Stamp: %s\n",
		serviceInfo.ServiceName, serviceInfo.ServiceStartTime, serviceInfo.GoVersion, serviceInfo.TimeStamp)

	go StartDashboard(addr)
}

func StartDashboard(addr int) {

	log.Println("Starting the dashboard")

	http.HandleFunc("/", serveHtmlSite)
	http.HandleFunc("/metrics", getMetrics)
	http.HandleFunc("/function-metrics", getFunctionMetrics)
	http.HandleFunc("/generate-function-metrics", profileHandler)

	// API to fetch the service metrics
	http.HandleFunc("/get-service-info", api.GetServiceInfoAPI)
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
		serviceInfo.ServiceName,
		serviceStartTime.Format(time.RFC3339),
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

func profileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Generating profile\n")
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	profilesFolderPath := fmt.Sprintf("%s/profiles", BasePath)

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

func GetBasePath() string {

	const monigoFolder string = "monigo"

	var path string
	appPath, _ := os.Getwd()
	if appPath == "/" {
		path = fmt.Sprintf("%s%s", appPath, monigoFolder)
	} else {
		path = fmt.Sprintf("%s/%s", appPath, monigoFolder)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	return path
}

func MeasureExecutionTime(name string, f func()) {
	core.MeasureExecutionTime(name, f)
}

func RecordRequestDuration(duration time.Duration) {
	core.RecordRequestDuration(duration)
}

func NewInitializeStorage() {
	timeseries.PurgeStorage()

	insertTimes := 150

	for i := 0; i < insertTimes; i++ {
		var serviceMetrics models.TimeSeriesServiceMetrics
		serviceMetrics.Load = float64(i)
		serviceMetrics.MemoryUsed = 0.21
		serviceMetrics.Cores = 0.21
		serviceMetrics.NumberOfReqServerd = 0.21
		serviceMetrics.GoRoutines = 0.21
		serviceMetrics.TotalAlloc = 0.21
		serviceMetrics.MemoryAllocSys = 0.21
		serviceMetrics.HeapAlloc = 0.21
		serviceMetrics.HeapAllocSys = 0.21
		serviceMetrics.UpTime = time.Duration(0)
		serviceMetrics.TotalDurationTookByAPI = time.Duration(0)

		if err := timeseries.StoreServiceMetrics(&serviceMetrics); err != nil {
			log.Fatalf("Error storing service metrics: %v\n", err)
		}

		log.Printf("Stored service metrics %d\n", i)
		time.Sleep(1 * time.Second)
	}

	time.Sleep(5 * time.Second)

	timestamp := time.Now()
	timestampInt := timestamp.Add(-24 * time.Hour).Unix()

	load, err := timeseries.GetDataPoints("load_metrics", []tstorage.Label{{Name: "host", Value: "server1"}}, timestampInt, timestamp.Unix())
	if err != nil {
		log.Fatalf("Error getting data points: %v\n", err)
	}

	jsonLoad, err := json.Marshal(load)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v\n", err)
	}

	log.Printf("Load: %s\n", string(jsonLoad))

	timeseries.CloseStorage() // Close storage only once, at the end
}
