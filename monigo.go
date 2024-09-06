package monigo

import (
	"embed"
	"fmt"
	"log"
	"net/http"
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
	Start()                                                                // Start the dashboard
	DeleteMonigoStorage()                                                  // Purge the monigo storage
	SetDbSyncFrequency(frequency ...string)                                // Set the frequency to sync the metrics to the storage
	PrintGoRoutinesStats() models.GoRoutinesStatistic                      // Print the Go routines stats
	SetServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) // Set the service thresholds to calculate the overall service health
}

func (m *Monigo) Start() {

	if m.ServiceName == "" {
		log.Panic("service_name is required, please provide the service name")
	}

	if m.PurgeMonigoStorage {
		log.Panic("PurgeMonigoStorage is set to true, please set it to false to start the service")
		m.DeleteMonigoStorage()
	}

	// Set the frequency to sync the metrics to the storage
	m.SetDbSyncFrequency(m.DbSyncFrequency) // Default is 5 Minutes

	//@TODO:  RetentionPeriod  Yet to be implemented

	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()
	m.ServiceStartTime = serviceStartTime

	common.SetServiceInfo(m.ServiceName, m.ServiceStartTime, m.GoVersion, m.ProcessId)
	go StartDashboard(m.DashboardPort)
}

func (m *Monigo) DeleteMonigoStorage() {
	log.Panic("PurgeMonigoStorage is set to true, please set it to false to start the service")
	timeseries.PurgeStorage()
}

func (m *Monigo) SetDbSyncFrequency(frequency ...string) {
	timeseries.SetDbSyncFrequency(m.DbSyncFrequency)
}

func (m *Monigo) PrintGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

func (m *Monigo) SetServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	core.SetServiceThresholds(thresholdsValues)
}

func StartDashboard(addr int) {

	if addr == 0 {
		addr = 8080 // Default port for the dashboard
	}

	log.Println("Starting the dashboard at port:", addr)

	http.HandleFunc("/", serveHtmlSite)
	http.HandleFunc("/metrics", api.NewCoreStatistics)

	http.HandleFunc("/function-metrics", api.GetFunctionMetrics)

	// http.HandleFunc("/generate-function-metrics", api.ProfileHandler)

	// API to fetch the service metrics
	http.HandleFunc("/service-info", api.GetServiceInfoAPI) // Completed

	http.HandleFunc("/service-metrics", api.GetServiceMetricsFromStorage) // API to fetch DATA points
	http.HandleFunc("/go-routines-stats", api.GetGoRoutinesStats)

	// /get-metrics?fields=service-info
	// http.HandleFunc("/get-metrics", api.GetMetricsInfo)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", addr), nil); err != nil {
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
		filePath = baseDir + "/assets/images/favicon.ico"
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

func MeasureExecutionTime(name string, f func()) {
	core.MeasureExecutionTime(name, f)
}

func RecordRequestDuration(duration time.Duration) {
	core.RecordRequestDuration(duration)
}
