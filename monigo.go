package monigo

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
	router      interface{}                      // Router to register the dashboard routes
	routePath   string             = "/"         // Route path for the dashboard
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

func RegisterDashboardRoute(routerType interface{}, rPath string) {

	// Supported Routers are mentioned below:
	// 1. *http.ServeMux
	// 2. *gin.Engine 	(In Progress)
	// 3. *mux.Router 	(In Progress)
	// 4. *chi.Mux 		(In Progress)
	// 5. *echo.Echo 	(In Progress)
	// 6. *gorilla.Mux 	(In Progress)
	// If the router is not provided, then the default router *http.ServeMux will be used

	// In case you want to need a router support for any other router
	// please raise a feature request on the GitHub repository, I will try to add the support for the same in the next release
	// Thanks, Yash!

	switch r := routerType.(type) {
	case *http.ServeMux:
		router = r
		routePath = rPath

	case *gin.Engine:
		router = r
		routePath = rPath

	default:
		log.Println("[MoniGo] Invalid router type. Supported types are *http.ServeMux and *gin.Engine, Setting to default router *http.ServeMux")
	}
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
	if routePath == "/" {
		setDashboardPort(m) // Set the dashboard port
	}
	startDashboard(router, routePath, m.DashboardPort) // Start the dashboard
}

// GetGoRoutinesStats get back the Go routines stats from the core package
func (m *Monigo) GetGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

// TraceFunction traces the function
func TraceFunction(f func()) {
	core.TraceFunction(f)
}

// startDashboard starts the dashboard on the specified port
func startDashboard(router interface{}, path string, port int) {
	if port != 0 && router == nil {
		log.Println("[MoniGo] Port and router not provided. Setting to default port and router (Default Port: 8080 and Default Router: *http.ServeMux)")
		r := http.NewServeMux()
		r.HandleFunc("/", serveStaticFiles("static", "/"))                                             // Serve the HTML site
		r.HandleFunc(fmt.Sprintf("%s/metrics", baseAPIPath), api.GetServiceStatistics)                 // Service Statistics API
		r.HandleFunc(fmt.Sprintf("%s/service-info", baseAPIPath), api.GetServiceInfoAPI)               // Service API to get the service information
		r.HandleFunc(fmt.Sprintf("%s/service-metrics", baseAPIPath), api.GetServiceMetricsFromStorage) // Service API to get the service metrics
		r.HandleFunc(fmt.Sprintf("%s/go-routines-stats", baseAPIPath), api.GetGoRoutinesStats)         // Service API to get the Go routines stats
		r.HandleFunc(fmt.Sprintf("%s/function", baseAPIPath), api.GetFunctionTraceDetails)             // Service API to get the function trace details
		r.HandleFunc(fmt.Sprintf("%s/function-details", baseAPIPath), api.ViewFunctionMaetrtics)       // Service API to get the function metrics
		r.HandleFunc(fmt.Sprintf("%s/reports", baseAPIPath), api.GetReportData)                        // Reports API to get the reports
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	} else {
		switch router := router.(type) {
		case *http.ServeMux:
			log.Println("[MoniGo] Starting the dashboard on the provided router path and port:", path)
			router.HandleFunc("/", serveStaticFiles("static", path)) // Serve the HTML site
			router.HandleFunc(fmt.Sprintf("%s/metrics", baseAPIPath), api.GetServiceStatistics)
			router.HandleFunc(fmt.Sprintf("%s/service-info", baseAPIPath), api.GetServiceInfoAPI)
			router.HandleFunc(fmt.Sprintf("%s/service-metrics", baseAPIPath), api.GetServiceMetricsFromStorage)
			router.HandleFunc(fmt.Sprintf("%s/go-routines-stats", baseAPIPath), api.GetGoRoutinesStats)
			router.HandleFunc(fmt.Sprintf("%s/function", baseAPIPath), api.GetFunctionTraceDetails)
			router.HandleFunc(fmt.Sprintf("%s/function-details", baseAPIPath), api.ViewFunctionMaetrtics)
			router.HandleFunc(fmt.Sprintf("%s/reports", baseAPIPath), api.GetReportData)
		case *gin.Engine:
			log.Println("[MoniGo] Starting the dashboard on the provided router: ", router)
			// @TODO: Need to work on the gin router
		default:
			log.Panic("[MoniGo] Invalid router type. Supported types are *http.ServeMux and *gin.Engine")
		}
	}
}

// serveStaticFiles serves HTML, CSS, JS, and other static files from the specified base directory.
func serveStaticFiles(baseDir, basePath string) http.HandlerFunc {
	contentTypes := map[string]string{ // Mapping of file extensions to content types
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

	return func(w http.ResponseWriter, r *http.Request) {
		trimmedPath := strings.TrimPrefix(r.URL.Path, basePath)
		if trimmedPath == "" || trimmedPath == "/" {
			trimmedPath = "/index.html"
		} else if trimmedPath == "/favicon.ico" {
			trimmedPath = "/assets/favicon.ico"
		}

		filePath := filepath.Join(baseDir, trimmedPath)
		ext := filepath.Ext(filePath)
		contentType := contentTypes[ext]
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		log.Println("[MoniGo] Requested path:", r.URL.Path, ", File path:", filePath)

		file, err := staticFiles.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Could not load "+filePath, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(file)
	}
}
