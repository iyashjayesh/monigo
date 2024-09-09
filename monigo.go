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
	staticFiles      embed.FS
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	BasePath         string
)

func init() {
	BasePath = common.GetBasePath()
}

// Monigo is the main struct to start the monigo service
type Monigo struct {
	ServiceName        string    `json:"service_name"`
	DashboardPort      int       `json:"dashboard_port"`
	PurgeMonigoStorage bool      `json:"purge_monigo_storage"`
	DbSyncFrequency    string    `json:"db_sync_frequency"`
	RetentionPeriod    string    `json:"retention_period"`
	GoVersion          string    `json:"go_version"`
	ServiceStartTime   time.Time `json:"service_start_time"`
	ProcessId          int32     `json:"process_id"`
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start()                                                                      // Start the dashboard
	DeleteMonigoStorage()                                                        // Purge the monigo storage
	SetDbSyncFrequency(frequency ...string)                                      // Set the frequency to sync the metrics to the storage
	PrintGoRoutinesStats() models.GoRoutinesStatistic                            // Print the Go routines stats
	ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) // Set the service thresholds to calculate the overall service health
}

type Cache struct {
	Data map[string]time.Time
}

func (m *Monigo) Start() {

	if m.ServiceName == "" {
		log.Panic("service_name is required, please provide the service name")
	}

	if m.PurgeMonigoStorage {
		m.DeleteMonigoStorage()
	}

	// Set the frequency to sync the metrics to the storage
	m.SetDbSyncFrequency(m.DbSyncFrequency) // Default is 5 Minutes

	// TODO: Correct the logs

	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()
	m.ServiceStartTime = serviceStartTime

	cachePath := BasePath + "/cache.dat"
	cache := Cache{Data: make(map[string]time.Time)}

	err := cache.LoadFromFile(cachePath)
	if err != nil {
		log.Println("Could not load cache, starting fresh")
	}

	if _, ok := cache.Data[m.ServiceName]; ok {
		m.ServiceStartTime = cache.Data[m.ServiceName]
	}

	cache.Data[m.ServiceName] = m.ServiceStartTime

	log.Println("Service start time updated in cache, new start time:", m.ServiceStartTime)

	err = cache.SaveToFile(cachePath)
	if err != nil {
		log.Fatal(err)
	}

	common.SetServiceInfo(m.ServiceName, m.ServiceStartTime, m.GoVersion, m.ProcessId, m.RetentionPeriod)
	go StartDashboard(m.DashboardPort)
}

func (m *Monigo) DeleteMonigoStorage() {
	timeseries.PurgeStorage()
}

func (m *Monigo) SetDbSyncFrequency(frequency ...string) {
	timeseries.SetDbSyncFrequency(m.DbSyncFrequency)
}

func (m *Monigo) PrintGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

func (m *Monigo) ConfigureServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	core.ConfigureServiceThresholds(thresholdsValues)
}

func StartDashboard(port int) {

	if port == 0 {
		port = 8080 // Default port for the dashboard
	}

	log.Println("Starting the dashboard at port:", port)

	// Base API path
	basePath := "/monigo/api/v1"

	// HTML site
	http.HandleFunc("/", serveHtmlSite)

	// Core Statistics
	http.HandleFunc(fmt.Sprintf("%s/metrics", basePath), api.NewCoreStatistics)

	// Service APIs
	http.HandleFunc(fmt.Sprintf("%s/service-info", basePath), api.GetServiceInfoAPI)
	http.HandleFunc(fmt.Sprintf("%s/service-metrics", basePath), api.GetServiceMetricsFromStorage)
	http.HandleFunc(fmt.Sprintf("%s/go-routines-stats", basePath), api.GetGoRoutinesStats)

	// Reports
	http.HandleFunc(fmt.Sprintf("%s/reports", basePath), api.GetReportData)

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

func (c *Cache) LoadFromFile(filename string) error {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the Base64 encoded data from the file
	base64Data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// Decode the Base64 data
	jsonData, err := base64.StdEncoding.DecodeString(string(base64Data))
	if err != nil {
		return err
	}

	// Decode the JSON data into the cache
	return json.Unmarshal(jsonData, &c.Data)
}
