package monigo

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/api"
	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/core"
	"github.com/iyashjayesh/monigo/timeseries"
)

var (
	//go:embed static/assets/* static/index.html static/function-metrics.html static/reports.html static/go-routines-stats.html
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
	ServiceName      string
	DashboardPort    int
	GoVersion        string
	ServiceStartTime time.Time
}

type MonigoInt interface {
	PurgeMonigoStorage()                    // Purge the monigo storage
	SetDbSyncFrequency(frequency ...string) // Set the frequency to sync the metrics to the storage
	StartDashboard()                        // Start the dashboard
	PrintGoRoutinesStats() (int, []string)  // Print the Go routines stats
}

func (m *Monigo) StartDashboard() {

	common.SetServiceInfo(m.ServiceName, serviceStartTime, runtime.Version())

	m.GoVersion = runtime.Version()
	m.ServiceStartTime = serviceStartTime

	if m.DashboardPort == 0 {
		m.DashboardPort = 8080
	}

	go StartDashboard(m.DashboardPort)
}

func (m *Monigo) PurgeMonigoStorage() {
	timeseries.PurgeStorage()
}

func (m *Monigo) SetDbSyncFrequency(frequency ...string) {
	timeseries.SetDbSyncFrequency(frequency...)
}

func (m *Monigo) PrintGoRoutinesStats() (int, []string) {
	return core.CollectGoRoutinesInfo()
}

func StartDashboard(addr int) {

	log.Println("Starting the dashboard")

	http.HandleFunc("/", serveHtmlSite)
	http.HandleFunc("/metrics", api.GetMetrics)
	http.HandleFunc("/function-metrics", api.GetFunctionMetrics)
	http.HandleFunc("/generate-function-metrics", api.ProfileHandler)

	// API to fetch the service metrics
	http.HandleFunc("/service-info", api.GetServiceInfoAPI)
	http.HandleFunc("/service-metrics", api.GetServiceMetricsFromStorage)
	http.HandleFunc("/go-routines-stats", api.GetGoRoutinesStats)

	// /get-metrics?fields=service-info
	http.HandleFunc("/get-metrics", api.GetMetricsInfo)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", addr), nil); err != nil {
		log.Panicf("Error starting the dashboard: %v\n", err)
	}
}

func serveHtmlSite(w http.ResponseWriter, r *http.Request) {
	filePath := "static" + r.URL.Path
	var contentType string
	switch {
	case r.URL.Path == "/":
		filePath = "static/index.html"
		contentType = "text/html"
	case r.URL.Path == "/function-metrics.html":
		filePath = "static/function-metrics.html"
		contentType = "text/html"
	case r.URL.Path == "/reports.html":
		filePath = "static/reports.html"
		contentType = "text/html"
	case r.URL.Path == "/go-routines-stats.html":
		filePath = "static/go-routines-stats.html"
		contentType = "text/html"
	case r.URL.Path == "/favicon.ico":
		filePath = "static/assets/images/favicon.ico"
		contentType = "image/x-icon"
	case strings.HasPrefix(r.URL.Path, "/assets/css/"):
		contentType = "text/css"
	case strings.HasPrefix(r.URL.Path, "/assets/js/"):
		contentType = "application/javascript"
	case strings.HasSuffix(r.URL.Path, ".png"):
		contentType = "image/png"
	case strings.HasSuffix(r.URL.Path, ".jpg") || strings.HasSuffix(r.URL.Path, ".jpeg"):
		contentType = "image/jpeg"
	case strings.HasSuffix(r.URL.Path, ".svg"):
		contentType = "image/svg+xml"
	case strings.HasSuffix(r.URL.Path, ".woff") || strings.HasSuffix(r.URL.Path, ".woff2"):
		contentType = "font/woff"
	default:
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
