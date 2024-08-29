package monigo

import (
	"fmt"
	"log"
	"net/http"
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
	serviceStartTime time.Time = time.Now()
	Once             sync.Once = sync.Once{}
	BasePath         string
	serviceInfo      models.ServiceInfo
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

	serviceInfo.ServiceName = m.ServiceName
	serviceInfo.ServiceStartTime = serviceStartTime
	serviceInfo.GoVersion = runtime.Version()
	serviceInfo.TimeStamp = serviceStartTime

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
	log.Printf("Setting db sync frequency: %v\n", frequency)
	timeseries.SetDbSyncFrequency(frequency...)
}

func (m *Monigo) ShowMetrics() {
	log.Println("Showing the metrics")
	timeseries.ShowMetrics()
}

func StartDashboard(addr int) {

	log.Println("Starting the dashboard")

	http.HandleFunc("/", api.ServeHtmlSite)
	http.HandleFunc("/metrics", api.GetMetrics)
	http.HandleFunc("/function-metrics", api.GetFunctionMetrics)
	http.HandleFunc("/generate-function-metrics", api.ProfileHandler)

	// API to fetch the service metrics
	// http.HandleFunc("/service-metrics", GetServiceMetricsFromMonigoDbData)

	fmt.Printf("Starting dashboard at http://localhost:%d\n", addr)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", addr), nil); err != nil {
		log.Fatalf("Error starting the dashboard: %v\n", err)
	}
}

func MeasureExecutionTime(name string, f func()) {
	core.MeasureExecutionTime(name, f)
}

func RecordRequestDuration(duration time.Duration) {
	core.RecordRequestDuration(duration)
}
