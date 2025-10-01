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

	"github.com/gofiber/fiber/v2"
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
	ServiceName             string    `json:"service_name"`         // Mandatory field ex. "backend", "OrderAPI", "PaymentService", etc.
	DashboardPort           int       `json:"dashboard_port"`       // Default is 8080
	DataPointsSyncFrequency string    `json:"db_sync_frequency"`    // Default is 5 Minutes
	DataRetentionPeriod     string    `json:"retention_period"`     // Default is 7 Day
	TimeZone                string    `json:"time_zone"`            // Default is Local
	GoVersion               string    `json:"go_version"`           // Dynamically set from runtime.Version()
	ServiceStartTime        time.Time `json:"service_start_time"`   // Dynamically setting it based on the service start time
	ProcessId               int32     `json:"process_id"`           // Dynamically set from os.Getpid()
	MaxCPUUsage             float64   `json:"max_cpu_usage"`        // Default is 95%, You can set it to 100% if you want to monitor 100% CPU usage
	MaxMemoryUsage          float64   `json:"max_memory_usage"`     // Default is 95%, You can set it to 100% if you want to monitor 100% Memory usage
	MaxGoRoutines           int       `json:"max_go_routines"`      // Default is 100, You can set it to any number based on your service
	CustomBaseAPIPath       string    `json:"custom_base_api_path"` // Custom base API path for integration with existing routers
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start()                                         // Start the monigo service with dashboard
	Initialize()                                    // Initialize monigo without starting dashboard
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

// MonigoInstanceConstructorWithoutPort is the constructor for the Monigo struct without port binding
// This is used for router integration where we don't want MoniGo to bind to any port
func (m *Monigo) MonigoInstanceConstructorWithoutPort() {

	if m.TimeZone == "" { // Setting default TimeZone if not provided
		m.TimeZone = "Local"
	}

	location, err := time.LoadLocation(m.TimeZone) // Loading the time zone location
	if err != nil {
		log.Println("[MoniGo] Error loading timezone. Setting to Local, Error: ", err)
		location = time.Local
	}

	// Skip setDashboardPort for router integration
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

// Initialize initializes the monigo service without starting the dashboard
// This is useful when you want to integrate MoniGo with your existing HTTP server
func (m *Monigo) Initialize() {
	// Validate service name
	if m.ServiceName == "" {
		log.Panic("[MoniGo] service_name is required, please provide the service name")
	}

	m.MonigoInstanceConstructorWithoutPort() // Use constructor without port binding
	timeseries.PurgeStorage()                // Purge storage and set sync frequency for metrics
	if err := timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency); err != nil {
		log.Panic("[MoniGo] failed to set data points sync frequency: ", err)
	}

	// Fetching runtime details
	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := common.Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		log.Printf("[MoniGo] Warning: failed to load cache from file: %v. Starting with fresh cache.", err)
		// Continue with empty cache instead of panicking
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
		log.Printf("[MoniGo] Warning: failed to save cache to file: %v", err)
		// Continue without saving cache
	}

	// Setting common service information
	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	// Initialize storage to ensure it's available for API calls
	_, err := timeseries.GetStorageInstance()
	if err != nil {
		log.Printf("[MoniGo] Warning: failed to initialize storage: %v", err)
	}
}

// Function to start the monigo service
func (m *Monigo) Start() {
	// Validate service name
	if m.ServiceName == "" {
		log.Panic("[MoniGo] service_name is required, please provide the service name")
	}

	m.MonigoInstanceConstructor() // Use the original constructor with port binding
	timeseries.PurgeStorage()     // Purge storage and set sync frequency for metrics
	if err := timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency); err != nil {
		log.Panic("[MoniGo] failed to set data points sync frequency: ", err)
	}

	// Fetching runtime details
	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := common.Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		log.Printf("[MoniGo] Warning: failed to load cache from file: %v. Starting with fresh cache.", err)
		// Continue with empty cache instead of panicking
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
		log.Printf("[MoniGo] Warning: failed to save cache to file: %v", err)
		// Continue without saving cache
	}

	// Setting common service information
	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	if err := StartDashboardWithCustomPath(m.DashboardPort, m.CustomBaseAPIPath); err != nil {
		log.Panic("[MoniGo] error starting the dashboard: ", err)
	}
}

// GetGoRoutinesStats get back the Go routines stats from the core package
func (m *Monigo) GetGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

// TraceFunction traces the function
// This is the original function maintained for backward compatibility
func TraceFunction(f func()) {
	core.TraceFunction(f)
}

// TraceFunctionWithArgs traces a function with parameters and captures the metrics
// This function uses reflection to call functions with arbitrary signatures
// Example usage:
//
//	func processUser(userID string) { ... }
//	monigo.TraceFunctionWithArgs(processUser, "123")
func TraceFunctionWithArgs(f interface{}, args ...interface{}) {
	core.TraceFunctionWithArgs(f, args...)
}

// TraceFunctionWithReturn traces a function with parameters and return values
// Returns the first result of the function call (for backward compatibility)
// Example usage:
//
//	func calculateTotal(items []Item) int { ... }
//	result := monigo.TraceFunctionWithReturn(calculateTotal, items)
func TraceFunctionWithReturn(f interface{}, args ...interface{}) interface{} {
	return core.TraceFunctionWithReturn(f, args...)
}

// TraceFunctionWithReturns traces a function with parameters and return values
// Returns all results of the function call as a slice of interface{}
// Example usage:
//
//	func processData(data []byte) (Result, error) { ... }
//	results := monigo.TraceFunctionWithReturns(processData, data)
//	result := results[0].(Result)
//	err := results[1].(error)
func TraceFunctionWithReturns(f interface{}, args ...interface{}) []interface{} {
	return core.TraceFunctionWithReturns(f, args...)
}

// StartDashboard starts the dashboard on the specified port
func StartDashboard(port int) error {
	return StartDashboardWithCustomPath(port, baseAPIPath)
}

// StartDashboardWithCustomPath starts the dashboard on the specified port with a custom API path
func StartDashboardWithCustomPath(port int, customBaseAPIPath string) error {
	if port == 0 {
		port = 8080 // Default port for the dashboard
	}

	apiPath := baseAPIPath
	if customBaseAPIPath != "" {
		apiPath = customBaseAPIPath
	}

	// HTML site
	http.HandleFunc("/", serveHtmlSite)

	// API to get Service Statistics
	http.HandleFunc(fmt.Sprintf("%s/metrics", apiPath), api.GetServiceStatistics)

	// Service APIs
	http.HandleFunc(fmt.Sprintf("%s/service-info", apiPath), api.GetServiceInfoAPI)
	http.HandleFunc(fmt.Sprintf("%s/service-metrics", apiPath), api.GetServiceMetricsFromStorage)
	http.HandleFunc(fmt.Sprintf("%s/go-routines-stats", apiPath), api.GetGoRoutinesStats)
	http.HandleFunc(fmt.Sprintf("%s/function", apiPath), api.GetFunctionTraceDetails)
	http.HandleFunc(fmt.Sprintf("%s/function-details", apiPath), api.ViewFunctionMaetrtics)

	// Reports
	http.HandleFunc(fmt.Sprintf("%s/reports", apiPath), api.GetReportData)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("error starting the dashboard: %v", err)
	}

	return nil
}

// RegisterDashboardHandlers registers all dashboard handlers (both API and static files) to the provided HTTP mux
// This allows developers to integrate MoniGo dashboard into their existing HTTP server
func RegisterDashboardHandlers(mux *http.ServeMux, customBaseAPIPath ...string) {
	// Use the unified handler internally
	unifiedHandler := GetUnifiedHandler(customBaseAPIPath...)
	mux.Handle("/", http.HandlerFunc(unifiedHandler))
}

// RegisterAPIHandlers registers only the API handlers to the provided HTTP mux
// This is useful when developers want to handle static file serving themselves
func RegisterAPIHandlers(mux *http.ServeMux, customBaseAPIPath ...string) {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	// Register only API handlers
	mux.HandleFunc(fmt.Sprintf("%s/metrics", apiPath), api.GetServiceStatistics)
	mux.HandleFunc(fmt.Sprintf("%s/service-info", apiPath), api.GetServiceInfoAPI)
	mux.HandleFunc(fmt.Sprintf("%s/service-metrics", apiPath), api.GetServiceMetricsFromStorage)
	mux.HandleFunc(fmt.Sprintf("%s/go-routines-stats", apiPath), api.GetGoRoutinesStats)
	mux.HandleFunc(fmt.Sprintf("%s/function", apiPath), api.GetFunctionTraceDetails)
	mux.HandleFunc(fmt.Sprintf("%s/function-details", apiPath), api.ViewFunctionMaetrtics)
	mux.HandleFunc(fmt.Sprintf("%s/reports", apiPath), api.GetReportData)
}

// RegisterStaticHandlers registers only the static file handlers to the provided HTTP mux
// This is useful when developers want to handle API routing themselves
func RegisterStaticHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", serveHtmlSite)
}

// GetAPIHandlers returns a map of API handlers that can be registered to any HTTP router
// This provides maximum flexibility for integration with different router libraries
func GetAPIHandlers(customBaseAPIPath ...string) map[string]http.HandlerFunc {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	return map[string]http.HandlerFunc{
		fmt.Sprintf("%s/metrics", apiPath):           api.GetServiceStatistics,
		fmt.Sprintf("%s/service-info", apiPath):      api.GetServiceInfoAPI,
		fmt.Sprintf("%s/service-metrics", apiPath):   api.GetServiceMetricsFromStorage,
		fmt.Sprintf("%s/go-routines-stats", apiPath): api.GetGoRoutinesStats,
		fmt.Sprintf("%s/function", apiPath):          api.GetFunctionTraceDetails,
		fmt.Sprintf("%s/function-details", apiPath):  api.ViewFunctionMaetrtics,
		fmt.Sprintf("%s/reports", apiPath):           api.GetReportData,
	}
}

// GetStaticHandler returns the static file handler function
// This can be used to register static file serving to any HTTP router
func GetStaticHandler() http.HandlerFunc {
	return serveHtmlSite
}

// GetUnifiedHandler returns a unified handler that handles both API and static files
// This is the recommended way to integrate MoniGo with any HTTP router
func GetUnifiedHandler(customBaseAPIPath ...string) http.HandlerFunc {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiPath) {
			routeToAPIHandler(w, r, apiPath)
			return
		}

		serveHtmlSite(w, r)
	}
}

// GetFiberHandler returns a Fiber-compatible handler that handles both API and static files
// This is specifically designed for Fiber framework integration
func GetFiberHandler(customBaseAPIPath ...string) func(*fiber.Ctx) error {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	return func(c *fiber.Ctx) error {
		path := string(c.Request().URI().Path())
		if strings.HasPrefix(path, apiPath) {
			return routeToFiberAPIHandler(c, path, apiPath)
		}
		return serveFiberStaticFiles(c, path)
	}
}

// routeToAPIHandler routes API requests to the appropriate handler
func routeToAPIHandler(w http.ResponseWriter, r *http.Request, apiPath string) {
	path := r.URL.Path

	switch {
	case path == fmt.Sprintf("%s/metrics", apiPath):
		api.GetServiceStatistics(w, r)
	case path == fmt.Sprintf("%s/service-info", apiPath):
		api.GetServiceInfoAPI(w, r)
	case path == fmt.Sprintf("%s/service-metrics", apiPath):
		api.GetServiceMetricsFromStorage(w, r)
	case path == fmt.Sprintf("%s/go-routines-stats", apiPath):
		api.GetGoRoutinesStats(w, r)
	case path == fmt.Sprintf("%s/function", apiPath):
		api.GetFunctionTraceDetails(w, r)
	case path == fmt.Sprintf("%s/function-details", apiPath):
		api.ViewFunctionMaetrtics(w, r)
	case path == fmt.Sprintf("%s/reports", apiPath):
		api.GetReportData(w, r)
	default:
		http.NotFound(w, r)
	}
}

// routeToFiberAPIHandler routes API requests to the appropriate handler for Fiber
func routeToFiberAPIHandler(c *fiber.Ctx, path, apiPath string) error {
	switch {
	case path == fmt.Sprintf("%s/metrics", apiPath):
		return handleFiberAPI(c, api.GetServiceStatistics)
	case path == fmt.Sprintf("%s/service-info", apiPath):
		return handleFiberAPI(c, api.GetServiceInfoAPI)
	case path == fmt.Sprintf("%s/service-metrics", apiPath):
		return handleFiberAPI(c, api.GetServiceMetricsFromStorage)
	case path == fmt.Sprintf("%s/go-routines-stats", apiPath):
		return handleFiberAPI(c, api.GetGoRoutinesStats)
	case path == fmt.Sprintf("%s/function", apiPath):
		return handleFiberAPI(c, api.GetFunctionTraceDetails)
	case path == fmt.Sprintf("%s/function-details", apiPath):
		return handleFiberAPI(c, api.ViewFunctionMaetrtics)
	case path == fmt.Sprintf("%s/reports", apiPath):
		return handleFiberAPI(c, api.GetReportData)
	default:
		c.Status(404).SendString("Not Found")
		return nil
	}
}

// handleFiberAPI converts Fiber context to HTTP and calls the API handler
func handleFiberAPI(c *fiber.Ctx, handler func(http.ResponseWriter, *http.Request)) error {
	// Creating a response writer adapter
	respWriter := &fiberResponseWriter{c: c}

	// Getting the request body
	body := c.Request().Body()

	// Creating a proper HTTP request from Fiber context with body
	req, err := http.NewRequest(
		string(c.Request().Header.Method()),
		"http://localhost"+string(c.Request().URI().Path()),
		strings.NewReader(string(body)),
	)
	if err != nil {
		c.Status(500).SendString("Internal Server Error")
		return nil
	}

	// Copying headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Setting Content-Length header if body is not empty
	if len(body) > 0 {
		req.ContentLength = int64(len(body))
	}

	// Calling the original handler
	handler(respWriter, req)

	return nil
}

// serveFiberStaticFiles serves static files for Fiber
func serveFiberStaticFiles(c *fiber.Ctx, path string) error {
	baseDir := "static"

	// Mapping of content types based on file extensions
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

	filePath := baseDir + path
	if path == "/" {
		filePath = baseDir + "/index.html"
	} else if path == "/favicon.ico" {
		filePath = baseDir + "/assets/favicon.ico"
	}

	ext := filepath.Ext(filePath)
	contentType, ok := contentTypes[ext]
	if !ok {
		contentType = "application/octet-stream"
	}

	file, err := staticFiles.ReadFile(filePath)
	if err != nil {
		c.Status(404).SendString("File not found")
		return nil
	}

	c.Set("Content-Type", contentType)
	return c.Send(file)
}

// serveHtmlSite serves the HTML, CSS, JS, and other static files
func serveHtmlSite(w http.ResponseWriter, r *http.Request) {
	baseDir := "static"
	// Mapping of content types based on file extensions
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

	ext := filepath.Ext(filePath)
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

// fiberResponseWriter adapts Fiber context to http.ResponseWriter
type fiberResponseWriter struct {
	c      *fiber.Ctx
	header http.Header
}

func (w *fiberResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *fiberResponseWriter) Write(data []byte) (int, error) {
	// Setting headers before writing
	if w.header != nil {
		for key, values := range w.header {
			for _, value := range values {
				w.c.Set(key, value)
			}
		}
	}
	return w.c.Write(data)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.c.Status(statusCode)
}
