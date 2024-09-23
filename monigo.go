package monigo

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/api"
	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/models"
	"github.com/iyashjayesh/monigo/timeseries"
)

var (
	//go:embed static/*
	staticFiles embed.FS                         // Embedding the static files
	Once        sync.Once          = sync.Once{} // Ensures that the storage is initialized only once
	BasePath    string                           // Base path for the monigo
	baseAPIPath = "/monigo/api/v1"               // Base API path for the dashboard
)

func init() {
	BasePath = common.GetBasePath() // Get the base path for the monigo
}

// Monigo is the main struct to start the monigo service
type Monigo struct {
	ServiceName             string    `json:"service_name"`       // Mandatory field ex. "backend", "OrderAPI", "PaymentService", etc.
	DashboardPort           int       `json:"dashboard_port"`     // Default is 8080
	DataPointsSyncFrequency string    `json:"db_sync_frequency"`  // Default is 5 Minutes
	DataRetentionPeriod     string    `json:"retention_period"`   // Default is 7 Day
	TimeZone                string    `json:"time_zone"`          // Default is Local
	GoVersion               string    `json:"go_version"`         // Dynamically set from runtime.Version()
	ServiceStartTime        time.Time `json:"service_start_time"` // Dynamically setting it based on the service start time
	ProcessId               int32     `json:"process_id"`         // Dynamically set from os.Getpid()
	MaxCPUUsage             float64   `json:"max_cpu_usage"`      // Default is 95%, You can set it to 100% if you want to monitor 100% CPU usage
	MaxMemoryUsage          float64   `json:"max_memory_usage"`   // Default is 95%, You can set it to 100% if you want to monitor 100% Memory usage
	MaxGoRoutines           int       `json:"max_go_routines"`    // Default is 100, You can set it to any number based on your service
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start()                                         // Purge the monigo storage
	GetGoRoutinesStats() models.GoRoutinesStatistic // Print the Go routines stats
}

// Cache is the struct to store the cache data
type Cache struct {
	Data map[string]time.Time
}

// setDashboardPort sets the dashboard port
func setDashboardPort(m *Monigo) {
	defaultPort := 8080

	if m.DashboardPort < 1 || m.DashboardPort > 65535 { // Validating the port range and check if no port is provided
		if m.DashboardPort == 0 {
			log.Println("[MoniGo] Port not provided. Setting to default port:", defaultPort)
		} else {
			log.Println("[MoniGo] Invalid port provided. Setting to default port:", defaultPort)
		}
		m.DashboardPort = defaultPort
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort)) // Attempting to listen on the provided or default port
	if err != nil {
		log.Printf("[MoniGo] Port %d in use. Setting to default port: %d\n", m.DashboardPort, defaultPort)
		m.DashboardPort = defaultPort
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort))
		if err != nil {
			log.Panicf("[MoniGo] Failed to bind to default port %d: %v\n", defaultPort, err)
		}
	}
	defer listener.Close()
}

// MonigoInstanceConstructor is the constructor for the Monigo struct
func (m *Monigo) MonigoInstanceConstructor() {

	if m.TimeZone == "" { // Setting default TimeZone if not provided
		m.TimeZone = "Local"
	}

	location, err := time.LoadLocation(m.TimeZone) // Loading the time zone location
	if err != nil {
		log.Println("[MoniGo] Error loading timezone. Setting to Local, Error: ", err)
		location = time.Local
	}

	setDashboardPort(m) // Setting the dashboard port
	m.DataPointsSyncFrequency = common.DefaultIfEmpty(m.DataPointsSyncFrequency, "5m")
	m.DataRetentionPeriod = common.DefaultIfEmpty(m.DataRetentionPeriod, "7d")
	m.MaxCPUUsage = common.DefaultFloatIfZero(m.MaxCPUUsage, 95)
	m.MaxMemoryUsage = common.DefaultFloatIfZero(m.MaxMemoryUsage, 95)
	m.MaxGoRoutines = common.DefaultIntIfZero(m.MaxGoRoutines, 100)

	core.ConfigureServiceThresholds(&models.ServiceHealthThresholds{
		MaxCPUUsage:    m.MaxCPUUsage,
		MaxMemoryUsage: m.MaxMemoryUsage,
		MaxGoRoutines:  m.MaxGoRoutines,
	})

	m.ServiceStartTime = time.Now().In(location) // Setting the service start time
}

// Function to start the monigo service
func (m *Monigo) Start() {
	// Validate service name
	if m.ServiceName == "" {
		log.Panic("[MoniGo] service_name is required, please provide the service name")
	}

	m.MonigoInstanceConstructor()
	timeseries.PurgeStorage() // Purge storage and set sync frequency for metrics
	if err := timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency); err != nil {
		log.Panic("[MoniGo] failed to set data points sync frequency: ", err)
	}

	// Fetching runtime details
	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := common.Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		log.Panic("[MoniGo] failed to load cache from file: ", err)
	}

	// Updating the service start time in the cache
	if startTime, exists := cache.Data[m.ServiceName]; exists {
		m.ServiceStartTime = startTime
	} else {
		m.ServiceStartTime = time.Now()
		cache.Data[m.ServiceName] = m.ServiceStartTime
	}

	// Save the cache data to file
	if err := cache.SaveToFile(cachePath); err != nil {
		log.Panic("[MoniGo] error saving cache to file: ", err)
	}

	// Setting common service information
	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	if err := StartDashboard(m.DashboardPort); err != nil {
		log.Panic("[MoniGo] error starting the dashboard: ", err)
	}
}

// GetGoRoutinesStats get back the Go routines stats from the core package
func (m *Monigo) GetGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

// TraceFunction traces the function
func TraceFunction(f func()) {
	core.TraceFunction(f)
}

// StartDashboard starts the dashboard on the specified port
func StartDashboard(port int) error {

	if port == 0 {
		port = 8080 // Default port for the dashboard
	}

	// HTML site
	http.HandleFunc("/", serveHtmlSite)

	// API to get Service Statistics
	http.HandleFunc(fmt.Sprintf("%s/metrics", baseAPIPath), api.GetServiceStatistics)

	// Service APIs
	http.HandleFunc(fmt.Sprintf("%s/service-info", baseAPIPath), api.GetServiceInfoAPI)
	http.HandleFunc(fmt.Sprintf("%s/service-metrics", baseAPIPath), api.GetServiceMetricsFromStorage)
	http.HandleFunc(fmt.Sprintf("%s/go-routines-stats", baseAPIPath), api.GetGoRoutinesStats)
	http.HandleFunc(fmt.Sprintf("%s/function", baseAPIPath), api.GetFunctionTraceDetails)
	http.HandleFunc(fmt.Sprintf("%s/function-details", baseAPIPath), api.ViewFunctionMaetrtics)

	// Reports
	http.HandleFunc(fmt.Sprintf("%s/reports", baseAPIPath), api.GetReportData)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("error starting the dashboard: %v", err)
	}

	return nil
}

// serveHtmlSite serves the HTML, CSS, JS, and other static files
func serveHtmlSite(w http.ResponseWriter, r *http.Request) {
	baseDir := "static"
	// Map of content types based on file extensions
	contentTypes := map[string]string{
		".html":  "text/html",
		".ico":   "image/x-icon",
		".css":   "text/css",
		".js":    "application/javascript",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".svg":   "image/svg+xml",
		".woff":  "font/woff",
		".woff2": "font/woff2",
	}

	filePath := baseDir + r.URL.Path
	if r.URL.Path == "/" {
		filePath = baseDir + "/index.html"
	} else if r.URL.Path == "/favicon.ico" {
		filePath = baseDir + "/assets/favicon.ico"
	}

	ext := filepath.Ext(filePath) // getting the file extension
	contentType, ok := contentTypes[ext]
	if !ok {
		contentType = "application/octet-stream"
	}

	file, err := staticFiles.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Could not load "+filePath, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(file)
}
