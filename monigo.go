package monigo

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

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
	ServiceName      string
	DashboardPort    int
	GoVersion        string
	ServiceStartTime time.Time
}

type MonigoInt interface {
	PurgeMonigoStorage()                    // Purge the monigo storage
	SetDbSyncFrequency(frequency ...string) // Set the frequency to sync the metrics to the storage
	StartDashboard()                        // Start the dashboard
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

func StartDashboard(addr int) {

	log.Println("Starting the dashboard")

	http.HandleFunc("/", serveHtmlSite)
	http.HandleFunc("/metrics", api.GetMetrics)
	http.HandleFunc("/function-metrics", api.GetFunctionMetrics)
	http.HandleFunc("/generate-function-metrics", api.ProfileHandler)

	// API to fetch the service metrics
	http.HandleFunc("/service-info", api.GetServiceInfoAPI)
	http.HandleFunc("/service-metrics", api.GetServiceMetricsFromStorage)

	// /get-metrics?fields=service-info
	http.HandleFunc("/get-metrics", api.GetMetricsInfo)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", addr), nil); err != nil {
		log.Panicf("Error starting the dashboard: %v\n", err)
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
	} else if r.URL.Path == "/stylesheets/main.css" {
		file, err := staticFiles.ReadFile("static/stylesheets/main.css")
		if err != nil {
			http.Error(w, "Could not load main.css", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(file)
		return
	} else if r.URL.Path == "/js/main.js" {
		file, err := staticFiles.ReadFile("static/js/main.js")
		if err != nil {
			http.Error(w, "Could not load main.js", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(file)
		return
	}

	http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))).ServeHTTP(w, r)
}

func MeasureExecutionTime(name string, f func()) {
	core.MeasureExecutionTime(name, f)
}

func RecordRequestDuration(duration time.Duration) {
	core.RecordRequestDuration(duration)
}
