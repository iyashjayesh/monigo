package monigo

import (
	"embed"
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
	monigodb "github.com/iyashjayesh/monigo/monigoDb"
	bolt "go.etcd.io/bbolt"
)

var (
	//go:embed static/*
	staticFiles      embed.FS
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	Db               *bolt.DB
	BasePath         string
	serviceInfo      models.ServiceInfo
	dbObj            *monigodb.DBWrapper
	mu               sync.Mutex = sync.Mutex{}
)

func init() {
	BasePath = GetBasePath()
	dbObj = monigodb.GetDbInstance()
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

func PurgeMonigoDb() {
	monigodb.PurgeMonigoDbFile()
}

func SetDbSyncFrequency(intervalStr ...string) {
	monigodb.SetDbSyncFrequency(intervalStr...)
}

func MeasureExecutionTime(name string, f func()) {
	core.MeasureExecutionTime(name, f)
}

func RecordRequestDuration(duration time.Duration) {
	core.RecordRequestDuration(duration)
}
