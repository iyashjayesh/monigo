package monigo

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/iyashjayesh/monigo/timeseries"

	"github.com/iyashjayesh/monigo/api"
	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
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
	MaxCPUUsage             float64   `json:"max_cpu_usage"`      // Default is 80%
	MaxMemoryUsage          float64   `json:"max_memory_usage"`   // Default is 80%
	MaxGoRoutines           int       `json:"max_go_routines"`    // Default is 1000
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start()                                         // Purge the monigo storage
	GetGoRoutinesStats() models.GoRoutinesStatistic // Print the Go routines stats
	// ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) // Set the service thresholds to calculate the overall service health
}

// Cache is the struct to store the cache data
type Cache struct {
	Data map[string]time.Time
}

// MonigoInstanceConstructor is the constructor for the Monigo struct
func (m *Monigo) MonigoInstanceConstructor() {

	// Setting default TimeZone if not provided
	if m.TimeZone == "" {
		m.TimeZone = "Local"
	}

	// Loading the time zone location
	location, err := time.LoadLocation(m.TimeZone)
	if err != nil {
		log.Println("Error loading timezone. Setting to Local:", err)
		location = time.Local
	}

	// Setting default values
	m.DashboardPort = 8080
	m.DataPointsSyncFrequency = common.DefaultIfEmpty(m.DataPointsSyncFrequency, "5m")
	m.DataRetentionPeriod = common.DefaultIfEmpty(m.DataRetentionPeriod, "7d")
	m.MaxCPUUsage = common.DefaultFloatIfZero(m.MaxCPUUsage, 80)
	m.MaxMemoryUsage = common.DefaultFloatIfZero(m.MaxMemoryUsage, 80)
	m.MaxGoRoutines = common.DefaultIntIfZero(m.MaxGoRoutines, 1000)

	core.ConfigureServiceThresholds(&models.ServiceHealthThresholds{
		MaxCPUUsage:    m.MaxCPUUsage,
		MaxMemoryUsage: m.MaxMemoryUsage,
		MaxGoRoutines:  m.MaxGoRoutines,
	})

	m.ServiceStartTime = time.Now().In(location) // Setting the service start time
}

// Starting the monigo service with the provided configurations
func (m *Monigo) Start() {

	if m.ServiceName == "" {
		log.Panic("service_name is required, please provide the service name")
	}
	m.MonigoInstanceConstructor() // Initializing the Monigo instance

	// Purging storage and setting sync frequency for metrics
	timeseries.PurgeStorage()
	timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency)

	// Fetch runtime details
	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		log.Printf("Failed to load cache from file: %v", err)
	}

	// Update service start time from cache if present
	if startTime, exists := cache.Data[m.ServiceName]; exists {
		m.ServiceStartTime = startTime
	} else {
		cache.Data[m.ServiceName] = m.ServiceStartTime
	}

	// Save the cache data to file
	if err := cache.SaveToFile(cachePath); err != nil {
		log.Printf("Error saving cache to file: %v\n", err)
		log.Println("Proceeding without saving the cache")
	}

	// Set common service information
	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	// Start the dashboard in a separate goroutine
	go StartDashboard(m.DashboardPort)
}

// GetGoRoutinesStats get back the Go routines stats from the core package
func (m *Monigo) GetGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

// ConfigureServiceThresholds sets the service thresholds to calculate the overall service health
// func (m *Monigo) ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
// 	core.ConfigureServiceThresholds(thresholdsValues)
// }

// TraceFunction traces the function
func TraceFunction(f func()) {
	core.TraceFunction(f)
}

func StartDashboard(port int) {

	if port == 0 {
		port = 8080 // Default port for the dashboard
	}

	// HTML site
	http.HandleFunc("/", serveHtmlSite)

	// Core Statistics
	http.HandleFunc(fmt.Sprintf("%s/metrics", baseAPIPath), api.NewCoreStatistics)

	// Service APIs
	http.HandleFunc(fmt.Sprintf("%s/service-info", baseAPIPath), api.GetServiceInfoAPI)
	http.HandleFunc(fmt.Sprintf("%s/service-metrics", baseAPIPath), api.GetServiceMetricsFromStorage)
	http.HandleFunc(fmt.Sprintf("%s/go-routines-stats", baseAPIPath), api.GetGoRoutinesStats)
	http.HandleFunc(fmt.Sprintf("%s/function", baseAPIPath), api.GetFunctionTraceDetails)
	http.HandleFunc(fmt.Sprintf("%s/function-details", baseAPIPath), api.ViewFunctionMaetrtics)

	// Reports
	http.HandleFunc(fmt.Sprintf("%s/reports", baseAPIPath), api.GetReportData)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Panicf("Error starting the dashboard: %v\n", err)
	}
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

// SaveToFile saves the cache to a file
func (c *Cache) SaveToFile(filename string) error {
	jsonData, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}

	base64Data := base64.StdEncoding.EncodeToString(jsonData)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(base64Data)
	return err
}

// LoadFromFile loads the cache from a file, or starts fresh if the file does not exist or an error occurs
func (c *Cache) LoadFromFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	if stat, err := file.Stat(); err != nil {
		return fmt.Errorf("failed to stat cache file: %w", err)
	} else if stat.Size() == 0 {
		log.Println("Cache file is empty, starting fresh")
		return nil
	}

	base64Data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	jsonData, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	if err := json.Unmarshal(jsonData, &c.Data); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return nil
}
