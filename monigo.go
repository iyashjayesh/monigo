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

	"github.com/iyashjayesh/monigo/api"
	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/timeseries"
)

var (
	//go:embed static/*
	staticFiles embed.FS
	// serviceStartTime time.Time
	Once        sync.Once = sync.Once{}
	BasePath    string
	baseAPIPath = "/monigo/api/v1"
)

func init() {
	BasePath = common.GetBasePath()
}

// Monigo is the main struct to start the monigo service
type Monigo struct {
	ServiceName             string    `json:"service_name"`       // Mandatory field ex. "bachend", "OrderAPI", "PaymentService", etc.
	DashboardPort           int       `json:"dashboard_port"`     // Default is 8080
	DataPointsSyncFrequency string    `json:"db_sync_frequency"`  // Default is 5 Minutes
	DataRetentionPeriod     string    `json:"retention_period"`   // Default is 7 Day
	TimeZone                string    `json:"time_zone"`          // Default is Local
	GoVersion               string    `json:"go_version"`         // Dynamically set from runtime.Version()
	ServiceStartTime        time.Time `json:"service_start_time"` // Dynamically setting it based on the service start time
	ProcessId               int32     `json:"process_id"`         // Dynamically set from os.Getpid()

}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start()                                                                      // Purge the monigo storage
	SetDataPointsSyncFrequency(frequency ...string)                              // Set the frequency to sync the metrics to the storage
	PrintGoRoutinesStats() models.GoRoutinesStatistic                            // Print the Go routines stats
	ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) // Set the service thresholds to calculate the overall service health
}

type Cache struct {
	Data map[string]time.Time
}

// MonigoInstanceContructor is the constructor for the Monigo struct
func (m *Monigo) MonigoInstanceContructor() {

	if m.TimeZone == "" {
		m.TimeZone = "Local"
	}

	location, err := time.LoadLocation(m.TimeZone)
	if err != nil {
		log.Println("setting the default timezone to Local, error occurred:", err)
		location = time.Local
	}

	// Set the default values
	m.DashboardPort = 8080
	if m.DataPointsSyncFrequency == "" {
		m.DataPointsSyncFrequency = "5m"
	}
	if m.DataRetentionPeriod == "" {
		m.DataRetentionPeriod = "7d"
	}
	m.ServiceStartTime = time.Now().In(location)
}

func (m *Monigo) Start() {

	if m.ServiceName == "" {
		log.Panic("service_name is required, please provide the service name")
	}

	m.MonigoInstanceContructor()
	timeseries.PurgeStorage()                               // Purge the monigo storage to start fresh
	m.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency) // Set the frequency to sync the metrics to the storage

	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := Cache{Data: make(map[string]time.Time)}
	cache.LoadFromFile(cachePath)
	if _, ok := cache.Data[m.ServiceName]; ok {
		m.ServiceStartTime = cache.Data[m.ServiceName]
	}

	cache.Data[m.ServiceName] = m.ServiceStartTime

	log.Println("Service start time updated in cache, new start time:", m.ServiceStartTime)

	err := cache.SaveToFile(cachePath)
	if err != nil {
		log.Fatal(err)
	}

	common.SetServiceInfo(m.ServiceName, m.ServiceStartTime, m.GoVersion, m.ProcessId, m.DataRetentionPeriod)
	go StartDashboard(m.DashboardPort)
}

func (m *Monigo) SetDataPointsSyncFrequency(frequency ...string) {
	timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency)
}

func (m *Monigo) PrintGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

func (m *Monigo) ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	core.ConfigureServiceThresholds(thresholdsValues)
}

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

func (c *Cache) SaveToFile(filename string) error {
	// Encode the data as JSON
	jsonData, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}

	// Encode the JSON data as Base64
	base64Data := base64.StdEncoding.EncodeToString(jsonData)

	// Save the Base64 encoded data to the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(base64Data)
	return err
}

// handleError is a helper function to log the error and panic
func handleError(msg string, err error) {
	log.Panicf("%s: %v", msg, err)
}

// LoadFromFile loads the cache from a file, or starts fresh if the file does not exist or an error occurs
func (c *Cache) LoadFromFile(filename string) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		handleError("Could not open or create cache file", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		handleError("Could not retrieve file info", err)
	}
	if fileInfo.Size() == 0 {
		log.Println("Cache file is empty, starting fresh")
		return
	}

	base64Data, err := ioutil.ReadAll(file)
	if err != nil {
		handleError("Could not read cache file", err)
	}

	jsonData, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		handleError("Could not decode cache file", err)
	}

	err = json.Unmarshal(jsonData, &c.Data)
	if err != nil {
		handleError("Could not unmarshal cache data", err)
	}
}
