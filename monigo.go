package monigo

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
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
	Headless                bool      `json:"headless"`             // If true, dashboard won't be started
	SamplingRate            int       `json:"sampling_rate"`        // Trace 1 in N calls
	StorageType             string    `json:"storage_type"`         // "disk" or "memory"

	// Security and Middleware Configuration
	DashboardMiddleware []func(http.Handler) http.Handler `json:"-"` // Middleware chain for dashboard access (static files)
	APIMiddleware       []func(http.Handler) http.Handler `json:"-"` // Middleware chain for API endpoints
	AuthFunction        func(*http.Request) bool          `json:"-"` // Simple authentication function for dashboard access
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start() error                                   // Start the monigo service with dashboard
	Initialize() error                              // Initialize monigo without starting dashboard
	GetGoRoutinesStats() models.GoRoutinesStatistic // Print the Go routines stats
}

// Cache is the struct to store the cache data
type Cache struct {
	Data map[string]time.Time
}

// setDashboardPort sets the dashboard port
func setDashboardPort(m *Monigo) error {
	defaultPort := 8080

	// If the port is not provided or is out of range, we will set it to the default port
	if m.DashboardPort <= 0 || m.DashboardPort > 65535 {
		log.Println("[MoniGo] Port not provided. Setting to default port:", defaultPort)
		m.DashboardPort = defaultPort
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort)) // Attempting to listen on the provided or default port
	if err != nil {
		// If the port is in use, we will set it to the default port
		if portInUse := m.isAddrInUse(err); portInUse {
			log.Printf("[MoniGo] Port %d in use. Setting to default port: %d\n", m.DashboardPort, defaultPort)
			m.DashboardPort = defaultPort

			// Attempting to listen on the default port
			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort))
			if err != nil {
				return fmt.Errorf("[MoniGo] Failed to bind to default port %d: %v", defaultPort, err)
			}
		}
	}
	defer listener.Close()
	return nil
}

// isAddrInUse checks if the error is due to address in use
func (m *Monigo) isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr *os.SyscallError
		return errors.As(opErr.Err, &sysErr) && errors.Is(sysErr.Err, syscall.EADDRINUSE)
	}
	return false
}

// GetRuningPort returns the running port
func (m *Monigo) GetRuningPort() int {
	return m.DashboardPort
}

// MonigoInstanceConstructor is the constructor for the Monigo struct
func (m *Monigo) MonigoInstanceConstructor() error {

	if m.TimeZone == "" { // Setting default TimeZone if not provided
		m.TimeZone = "Local"
	}

	location, err := time.LoadLocation(m.TimeZone) // Loading the time zone location
	if err != nil {
		log.Println("[MoniGo] Error loading timezone. Setting to Local, Error: ", err)
		location = time.Local
	}

	if err := setDashboardPort(m); err != nil {
		return err
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
	return nil
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

// setup contains common initialization logic for both Initialize and Start
func (m *Monigo) setup() error {
	// Validate service name
	if m.ServiceName == "" {
		return fmt.Errorf("[MoniGo] service_name is required, please provide the service name")
	}

	if err := timeseries.PurgeStorage(); err != nil {
		return fmt.Errorf("[MoniGo] Warning: failed to purge storage: %v", err)
	}
	if err := timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency); err != nil {
		return fmt.Errorf("[MoniGo] failed to set data points sync frequency: %v", err)
	}

	// Fetching runtime details
	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := common.Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		log.Printf("[MoniGo] Warning: failed to load cache from file: %v. Starting with fresh cache.", err)
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
	}

	// Setting common service information
	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	// Initialize storage and core settings
	if m.StorageType != "" {
		timeseries.SetStorageType(m.StorageType)
	}
	if m.SamplingRate > 0 {
		core.SetSamplingRate(m.SamplingRate)
	}

	_, err := timeseries.GetStorageInstance()
	if err != nil {
		log.Printf("[MoniGo] Warning: failed to initialize storage: %v", err)
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	return nil
}

// Initialize initializes the monigo service without starting the dashboard
func (m *Monigo) Initialize() error {
	m.MonigoInstanceConstructorWithoutPort()
	return m.setup()
}

// Start starts the monigo service with dashboard
func (m *Monigo) Start() error {
	if err := m.MonigoInstanceConstructor(); err != nil {
		return err
	}

	if err := m.setup(); err != nil {
		return err
	}

	if m.Headless {
		log.Println("[MoniGo] Running in headless mode. Dashboard disabled.")
		return nil
	}

	if err := StartDashboardWithCustomPath(m.DashboardPort, m.CustomBaseAPIPath); err != nil {
		return fmt.Errorf("[MoniGo] error starting the dashboard: %v", err)
	}
	return nil
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

// SetSamplingRate sets the sampling rate for function tracing
// For example, a rate of 100 means 1 in 100 calls will be profiled
func SetSamplingRate(rate int) {
	core.SetSamplingRate(rate)
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
	http.HandleFunc("/metrics", api.PrometheusMetricsHandler)

	// Reports
	http.HandleFunc(fmt.Sprintf("%s/reports", apiPath), api.GetReportData)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return fmt.Errorf("error starting the dashboard: %v", err)
	}

	return nil
}

// StartSecuredDashboard starts the dashboard with middleware support
func StartSecuredDashboard(m *Monigo) error {
	if m.DashboardPort == 0 {
		m.DashboardPort = 8080 // Default port for the dashboard
	}

	// Get secured handlers
	unifiedHandler := GetSecuredUnifiedHandler(m, m.CustomBaseAPIPath)

	// Register the secured handler
	http.HandleFunc("/", unifiedHandler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", m.DashboardPort), nil); err != nil {
		return fmt.Errorf("error starting the secured dashboard: %v", err)
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

// RegisterSecuredDashboardHandlers registers all dashboard handlers with middleware support to the provided HTTP mux
// This allows developers to integrate MoniGo dashboard with security middleware into their existing HTTP server
func RegisterSecuredDashboardHandlers(mux *http.ServeMux, m *Monigo, customBaseAPIPath ...string) {
	// Use the secured unified handler internally
	unifiedHandler := GetSecuredUnifiedHandler(m, customBaseAPIPath...)
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
	mux.HandleFunc(fmt.Sprintf("%s/metrics", apiPath), api.GetServiceStatistics) // Existing JSON metrics
	mux.HandleFunc("/metrics", api.PrometheusMetricsHandler)                     // New Prometheus metrics
	mux.HandleFunc(fmt.Sprintf("%s/reports", apiPath), api.GetReportData)
}

// RegisterSecuredAPIHandlers registers only the API handlers with middleware support to the provided HTTP mux
// This is useful when developers want to handle static file serving themselves but need API security
func RegisterSecuredAPIHandlers(mux *http.ServeMux, m *Monigo, customBaseAPIPath ...string) {
	// Get secured API handlers
	securedHandlers := GetSecuredAPIHandlers(m, customBaseAPIPath...)

	// Register secured API handlers
	for path, handler := range securedHandlers {
		mux.HandleFunc(path, handler)
	}
}

// RegisterStaticHandlers registers only the static file handlers to the provided HTTP mux
// This is useful when developers want to handle API routing themselves
func RegisterStaticHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", serveHtmlSite)
}

// RegisterSecuredStaticHandlers registers only the static file handlers with middleware support to the provided HTTP mux
// This is useful when developers want to handle API routing themselves but need static file security
func RegisterSecuredStaticHandlers(mux *http.ServeMux, m *Monigo) {
	securedHandler := GetSecuredStaticHandler(m)
	mux.HandleFunc("/", securedHandler)
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
		"/metrics":                         api.PrometheusMetricsHandler,
		fmt.Sprintf("%s/reports", apiPath): api.GetReportData,
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

// GetSecuredUnifiedHandler returns a unified handler with middleware support for both API and static files
// This is the recommended way to integrate MoniGo with security middleware
func GetSecuredUnifiedHandler(m *Monigo, customBaseAPIPath ...string) http.HandlerFunc {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	// Create the base handler
	baseHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiPath) {
			routeToAPIHandler(w, r, apiPath)
			return
		}
		serveHtmlSite(w, r)
	}

	// Apply middleware if provided
	return applyMiddlewareChain(baseHandler, m.DashboardMiddleware, m.AuthFunction)
}

// GetSecuredAPIHandlers returns a map of API handlers with middleware support
func GetSecuredAPIHandlers(m *Monigo, customBaseAPIPath ...string) map[string]http.HandlerFunc {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	baseHandlers := map[string]http.HandlerFunc{
		fmt.Sprintf("%s/metrics", apiPath):           api.GetServiceStatistics,
		fmt.Sprintf("%s/service-info", apiPath):      api.GetServiceInfoAPI,
		fmt.Sprintf("%s/service-metrics", apiPath):   api.GetServiceMetricsFromStorage,
		fmt.Sprintf("%s/go-routines-stats", apiPath): api.GetGoRoutinesStats,
		fmt.Sprintf("%s/function", apiPath):          api.GetFunctionTraceDetails,
		fmt.Sprintf("%s/function-details", apiPath):  api.ViewFunctionMaetrtics,
		"/metrics":                         api.PrometheusMetricsHandler,
		fmt.Sprintf("%s/reports", apiPath): api.GetReportData,
	}

	// Apply middleware to each handler
	securedHandlers := make(map[string]http.HandlerFunc)
	for path, handler := range baseHandlers {
		securedHandlers[path] = applyMiddlewareChain(handler, m.APIMiddleware, nil)
	}

	return securedHandlers
}

// GetSecuredStaticHandler returns the static file handler with middleware support
func GetSecuredStaticHandler(m *Monigo) http.HandlerFunc {
	return applyMiddlewareChain(serveHtmlSite, m.DashboardMiddleware, m.AuthFunction)
}

// applyMiddlewareChain applies a chain of middleware to a handler
func applyMiddlewareChain(handler http.HandlerFunc, middleware []func(http.Handler) http.Handler, authFunc func(*http.Request) bool) http.HandlerFunc {
	// Start with the base handler
	var finalHandler http.Handler = http.HandlerFunc(handler)

	// Apply authentication function if provided
	if authFunc != nil {
		finalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for static files
			if isStaticFile(r.URL.Path) {
				handler(w, r)
				return
			}

			if !authFunc(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			handler(w, r)
		})
	}

	// Apply middleware chain in reverse order (last middleware wraps the handler first)
	for i := len(middleware) - 1; i >= 0; i-- {
		finalHandler = middleware[i](finalHandler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalHandler.ServeHTTP(w, r)
	})
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

// Built-in Security Middleware Functions

// BasicAuthMiddleware creates a basic authentication middleware
// Usage: monigo.BasicAuthMiddleware("admin", "password")
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for static files
			if isStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="MoniGo Dashboard"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyMiddleware creates an API key authentication middleware
// Usage: monigo.APIKeyMiddleware("your-secret-api-key")
func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for static files
			if isStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			providedKey := r.Header.Get("X-API-Key")
			if providedKey == "" {
				providedKey = r.URL.Query().Get("api_key")
			}

			if providedKey != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IPWhitelistMiddleware creates an IP whitelist middleware
// Usage: monigo.IPWhitelistMiddleware([]string{"127.0.0.1", "192.168.1.0/24"})
func IPWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip IP check for static files
			if isStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := getClientIP(r)

			// Check if client IP is in whitelist
			for _, allowedIP := range allowedIPs {
				if isIPAllowed(clientIP, allowedIP) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}

// RateLimitMiddleware creates a simple rate limiting middleware
// Usage: monigo.RateLimitMiddleware(100, time.Minute) // 100 requests per minute
func RateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
	type clientInfo struct {
		count     int
		lastReset time.Time
	}

	var mu sync.Mutex
	var clients = make(map[string]*clientInfo)

	// Start a cleanup goroutine to prevent memory leak
	go func() {
		ticker := time.NewTicker(window * 2)
		for range ticker.C {
			mu.Lock()
			for ip, info := range clients {
				if time.Since(info.lastReset) > window*2 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			mu.Lock()
			client, exists := clients[clientIP]
			if !exists {
				client = &clientInfo{count: 0, lastReset: now}
				clients[clientIP] = client
			}

			// Reset counter if window has passed
			if now.Sub(client.lastReset) > window {
				client.count = 0
				client.lastReset = now
			}

			// Check if limit exceeded
			if client.count >= requests {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			client.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware creates a request logging middleware
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			log.Printf("[MoniGo] %s %s %d %v %s", r.Method, r.URL.Path, wrapped.statusCode, duration, r.RemoteAddr)
		})
	}
}

// Helper functions for middleware

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isIPAllowed checks if an IP is allowed based on CIDR notation or exact match
func isIPAllowed(clientIP, allowedIP string) bool {
	// Handle CIDR notation
	if strings.Contains(allowedIP, "/") {
		_, network, err := net.ParseCIDR(allowedIP)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return false
		}
		return network.Contains(ip)
	}

	// Exact match
	return clientIP == allowedIP
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// isStaticFile checks if the request path is for a static file
func isStaticFile(path string) bool {
	// List of static file extensions
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".map", ".json", ".xml",
		".pdf", ".zip", ".txt", ".html", ".htm",
	}

	// Check if path has a static file extension
	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	// Check for common static file paths
	staticPaths := []string{
		"/css/", "/js/", "/assets/", "/images/", "/fonts/", "/static/",
	}

	for _, staticPath := range staticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	// Special case for favicon
	if path == "/favicon.ico" {
		return true
	}

	return false
}
