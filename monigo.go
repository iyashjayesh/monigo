package monigo

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
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
	"github.com/iyashjayesh/monigo/exporters"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/models"
	"github.com/iyashjayesh/monigo/timeseries"
)

var (
	//go:embed static/*
	staticFiles embed.FS
	BasePath    string
	baseAPIPath = "/monigo/api/v1"

	// Content-type mapping shared by both HTTP and Fiber static file handlers.
	staticContentTypes = map[string]string{
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
)

func init() {
	BasePath = common.GetBasePath()
}

// Monigo is the main struct to start the monigo service
type Monigo struct {
	ServiceName             string    `json:"service_name"`
	DashboardPort           int       `json:"dashboard_port"`
	DataPointsSyncFrequency string    `json:"db_sync_frequency"`
	DataRetentionPeriod     string    `json:"retention_period"`
	TimeZone                string    `json:"time_zone"`
	GoVersion               string    `json:"go_version"`
	ServiceStartTime        time.Time `json:"service_start_time"`
	ProcessId               int32     `json:"process_id"`
	MaxCPUUsage             float64   `json:"max_cpu_usage"`
	MaxMemoryUsage          float64   `json:"max_memory_usage"`
	MaxGoRoutines           int       `json:"max_go_routines"`
	CustomBaseAPIPath       string    `json:"custom_base_api_path"`
	Headless                bool      `json:"headless"`
	SamplingRate            int       `json:"sampling_rate"`
	StorageType             string    `json:"storage_type"`

	// OpenTelemetry Configuration
	OTelEndpoint string            `json:"otel_endpoint,omitempty"`
	OTelHeaders  map[string]string `json:"-"`

	// Security and Middleware Configuration
	DashboardMiddleware []func(http.Handler) http.Handler `json:"-"`
	APIMiddleware       []func(http.Handler) http.Handler `json:"-"`
	AuthFunction        func(*http.Request) bool          `json:"-"`

	// Holds a reference so we can shut down cleanly.
	otelExporter *exporters.OTelExporter
}

// MonigoInt is the interface to start the monigo service
type MonigoInt interface {
	Start() error
	Initialize() error
	GetGoRoutinesStats() models.GoRoutinesStatistic
}

// Cache is the struct to store the cache data
type Cache struct {
	Data map[string]time.Time
}

// setDashboardPort validates and binds the dashboard port.
func setDashboardPort(m *Monigo) error {
	defaultPort := 8080

	if m.DashboardPort <= 0 || m.DashboardPort > 65535 {
		logger.Log.Info("port not provided or out of range, setting to default", "port", defaultPort)
		m.DashboardPort = defaultPort
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort))
	if err != nil {
		if portInUse := m.isAddrInUse(err); portInUse {
			logger.Log.Warn("port in use, setting to default", "requested", m.DashboardPort, "default", defaultPort)
			m.DashboardPort = defaultPort

			listener, err = net.Listen("tcp", fmt.Sprintf(":%d", m.DashboardPort))
			if err != nil {
				return fmt.Errorf("[MoniGo] Failed to bind to default port %d: %v", defaultPort, err)
			}
		} else {
			return fmt.Errorf("[MoniGo] Failed to bind to port %d: %v", m.DashboardPort, err)
		}
	}
	defer listener.Close()
	return nil
}

func (m *Monigo) isAddrInUse(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var sysErr *os.SyscallError
		return errors.As(opErr.Err, &sysErr) && errors.Is(sysErr.Err, syscall.EADDRINUSE)
	}
	return false
}

// GetRunningPort returns the running port
func (m *Monigo) GetRunningPort() int {
	return m.DashboardPort
}

func (m *Monigo) initCommon() {
	if m.TimeZone == "" {
		m.TimeZone = "Local"
	}

	location, err := time.LoadLocation(m.TimeZone)
	if err != nil {
		logger.Log.Warn("error loading timezone, using Local", "error", err)
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

	m.ServiceStartTime = time.Now().In(location)
}

// MonigoInstanceConstructor validates the port then initialises common fields.
func (m *Monigo) MonigoInstanceConstructor() error {
	if err := setDashboardPort(m); err != nil {
		return err
	}
	m.initCommon()
	return nil
}

// MonigoInstanceConstructorWithoutPort initialises common fields without port binding.
func (m *Monigo) MonigoInstanceConstructorWithoutPort() {
	m.initCommon()
}

// setup contains common initialization logic for both Initialize and Start
func (m *Monigo) setup() error {
	if m.ServiceName == "" {
		return fmt.Errorf("[MoniGo] service_name is required, please provide the service name")
	}

	if err := timeseries.SetDataPointsSyncFrequency(m.DataPointsSyncFrequency); err != nil {
		return fmt.Errorf("[MoniGo] failed to set data points sync frequency: %v", err)
	}

	m.ProcessId = common.GetProcessId()
	m.GoVersion = runtime.Version()

	cachePath := BasePath + "/cache.dat"
	cache := common.Cache{Data: make(map[string]time.Time)}
	if err := cache.LoadFromFile(cachePath); err != nil {
		logger.Log.Warn("failed to load cache, starting fresh", "error", err)
	}

	if startTime, exists := cache.Data[m.ServiceName]; exists {
		m.ServiceStartTime = startTime
	} else {
		m.ServiceStartTime = time.Now()
		cache.Data[m.ServiceName] = m.ServiceStartTime
	}

	if err := cache.SaveToFile(cachePath); err != nil {
		logger.Log.Warn("failed to save cache", "error", err)
	}

	common.SetServiceInfo(
		m.ServiceName,
		m.ServiceStartTime,
		m.GoVersion,
		m.ProcessId,
		m.DataRetentionPeriod,
	)

	if m.StorageType != "" {
		timeseries.SetStorageType(m.StorageType)
	}
	if m.SamplingRate > 0 {
		core.SetSamplingRate(m.SamplingRate)
	}

	_, err := timeseries.GetStorageInstance()
	if err != nil {
		logger.Log.Error("failed to initialize storage", "error", err)
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	if m.OTelEndpoint != "" {
		otelExp, otelErr := exporters.NewOTelExporter(context.Background(), exporters.OTelConfig{
			Endpoint: m.OTelEndpoint,
			Headers:  m.OTelHeaders,
			Insecure: true,
		})
		if otelErr != nil {
			logger.Log.Error("failed to initialize OTel exporter", "error", otelErr)
		} else {
			m.otelExporter = otelExp
			logger.Log.Info("OTel exporter initialized", "endpoint", m.OTelEndpoint)
		}
	}

	return nil
}

// Shutdown performs a graceful cleanup of resources (OTel provider, storage, etc.).
func (m *Monigo) Shutdown(ctx context.Context) error {
	var errs []error
	if m.otelExporter != nil {
		if err := m.otelExporter.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("otel shutdown: %w", err))
		}
	}
	if err := timeseries.CloseStorage(); err != nil {
		errs = append(errs, fmt.Errorf("storage close: %w", err))
	}
	return errors.Join(errs...)
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
		logger.Log.Info("running in headless mode, dashboard disabled")
		return nil
	}

	if err := m.startDashboard(m.DashboardPort, m.CustomBaseAPIPath); err != nil {
		return fmt.Errorf("[MoniGo] error starting the dashboard: %v", err)
	}
	return nil
}

// GetGoRoutinesStats returns Go routines statistics.
func (m *Monigo) GetGoRoutinesStats() models.GoRoutinesStatistic {
	return core.CollectGoRoutinesInfo()
}

// TraceFunction traces the function
func TraceFunction(ctx context.Context, f func()) {
	core.TraceFunction(ctx, f)
}

// SetSamplingRate sets the sampling rate for function tracing
func SetSamplingRate(rate int) {
	core.SetSamplingRate(rate)
}

// TraceFunctionWithArgs traces a function with parameters and captures the metrics
func TraceFunctionWithArgs(ctx context.Context, f interface{}, args ...interface{}) {
	core.TraceFunctionWithArgs(ctx, f, args...)
}

// TraceFunctionWithReturn traces a function with parameters and return values
func TraceFunctionWithReturn(ctx context.Context, f interface{}, args ...interface{}) interface{} {
	return core.TraceFunctionWithReturn(ctx, f, args...)
}

// TraceFunctionWithReturns traces a function with parameters and returns all results
func TraceFunctionWithReturns(ctx context.Context, f interface{}, args ...interface{}) []interface{} {
	return core.TraceFunctionWithReturns(ctx, f, args...)
}

// StartDashboard starts the dashboard on the specified port
func StartDashboard(port int) error {
	m := &Monigo{}
	return m.startDashboard(port, baseAPIPath)
}

// StartDashboardWithCustomPath starts the dashboard on the specified port with a custom API path
func StartDashboardWithCustomPath(port int, customBaseAPIPath string) error {
	m := &Monigo{}
	return m.startDashboard(port, customBaseAPIPath)
}

func (m *Monigo) startDashboard(port int, customBaseAPIPath string) error {
	if port <= 0 || port > 65535 {
		port = 8080
	}

	apiPath := baseAPIPath
	if customBaseAPIPath != "" {
		apiPath = customBaseAPIPath
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHtmlSite)

	registerAPIEndpoints(mux, apiPath)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	m.registerShutdownHandler(srv)

	logger.Log.Info("dashboard started", "url", fmt.Sprintf("http://localhost:%d", port))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error starting the dashboard: %v", err)
	}

	return nil
}

// StartSecuredDashboard starts the dashboard with middleware support
func StartSecuredDashboard(m *Monigo) error {
	if m.DashboardPort <= 0 || m.DashboardPort > 65535 {
		m.DashboardPort = 8080
	}

	mux := http.NewServeMux()
	unifiedHandler := GetSecuredUnifiedHandler(m, m.CustomBaseAPIPath)
	mux.HandleFunc("/", unifiedHandler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", m.DashboardPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	m.registerShutdownHandler(srv)

	logger.Log.Info("secured dashboard started", "url", fmt.Sprintf("http://localhost:%d", m.DashboardPort))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error starting the secured dashboard: %v", err)
	}

	return nil
}

// registerShutdownHandler sets up a goroutine that listens for SIGINT/SIGTERM
// and performs a graceful server + storage shutdown.
func (m *Monigo) registerShutdownHandler(srv *http.Server) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Log.Info("shutting down dashboard server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Error("error during server shutdown", "error", err)
		}
		if err := m.Shutdown(ctx); err != nil {
			logger.Log.Error("error during resource cleanup", "error", err)
		}
	}()
}

// registerAPIEndpoints registers the standard API endpoints on the mux.
func registerAPIEndpoints(mux *http.ServeMux, apiPath string) {
	mux.HandleFunc(fmt.Sprintf("%s/metrics", apiPath), api.GetServiceStatistics)
	mux.HandleFunc(fmt.Sprintf("%s/service-info", apiPath), api.GetServiceInfoAPI)
	mux.HandleFunc(fmt.Sprintf("%s/service-metrics", apiPath), api.GetServiceMetricsFromStorage)
	mux.HandleFunc(fmt.Sprintf("%s/go-routines-stats", apiPath), api.GetGoRoutinesStats)
	mux.HandleFunc(fmt.Sprintf("%s/function", apiPath), api.GetFunctionTraceDetails)
	mux.HandleFunc(fmt.Sprintf("%s/function-details", apiPath), api.ViewFunctionMetrics)
	mux.HandleFunc("/metrics", api.PrometheusMetricsHandler)
	mux.HandleFunc(fmt.Sprintf("%s/reports", apiPath), api.GetReportData)
}

// RegisterDashboardHandlers registers all dashboard handlers to the provided HTTP mux
func RegisterDashboardHandlers(mux *http.ServeMux, customBaseAPIPath ...string) {
	unifiedHandler := GetUnifiedHandler(customBaseAPIPath...)
	mux.Handle("/", http.HandlerFunc(unifiedHandler))
}

// RegisterSecuredDashboardHandlers registers all dashboard handlers with middleware support
func RegisterSecuredDashboardHandlers(mux *http.ServeMux, m *Monigo, customBaseAPIPath ...string) {
	unifiedHandler := GetSecuredUnifiedHandler(m, customBaseAPIPath...)
	mux.Handle("/", http.HandlerFunc(unifiedHandler))
}

// RegisterAPIHandlers registers only the API handlers
func RegisterAPIHandlers(mux *http.ServeMux, customBaseAPIPath ...string) {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}
	registerAPIEndpoints(mux, apiPath)
}

// RegisterSecuredAPIHandlers registers only the API handlers with middleware support
func RegisterSecuredAPIHandlers(mux *http.ServeMux, m *Monigo, customBaseAPIPath ...string) {
	securedHandlers := GetSecuredAPIHandlers(m, customBaseAPIPath...)
	for path, handler := range securedHandlers {
		mux.HandleFunc(path, handler)
	}
}

// RegisterStaticHandlers registers only the static file handlers
func RegisterStaticHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", serveHtmlSite)
}

// RegisterSecuredStaticHandlers registers only the static file handlers with middleware
func RegisterSecuredStaticHandlers(mux *http.ServeMux, m *Monigo) {
	securedHandler := GetSecuredStaticHandler(m)
	mux.HandleFunc("/", securedHandler)
}

// GetAPIHandlers returns a map of API handlers for any HTTP router
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
		fmt.Sprintf("%s/function-details", apiPath):  api.ViewFunctionMetrics,
		"/metrics":                                   api.PrometheusMetricsHandler,
		fmt.Sprintf("%s/reports", apiPath):           api.GetReportData,
	}
}

// GetStaticHandler returns the static file handler function
func GetStaticHandler() http.HandlerFunc {
	return serveHtmlSite
}

// GetUnifiedHandler returns a unified handler that handles both API and static files
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

// GetFiberHandler returns a Fiber-compatible handler
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

// GetSecuredUnifiedHandler returns a unified handler with middleware
func GetSecuredUnifiedHandler(m *Monigo, customBaseAPIPath ...string) http.HandlerFunc {
	apiPath := baseAPIPath
	if len(customBaseAPIPath) > 0 && customBaseAPIPath[0] != "" {
		apiPath = customBaseAPIPath[0]
	}

	baseHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, apiPath) {
			routeToAPIHandler(w, r, apiPath)
			return
		}
		serveHtmlSite(w, r)
	}

	return applyMiddlewareChain(baseHandler, m.DashboardMiddleware, m.AuthFunction)
}

// GetSecuredAPIHandlers returns secured API handlers
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
		fmt.Sprintf("%s/function-details", apiPath):  api.ViewFunctionMetrics,
		"/metrics":                                   api.PrometheusMetricsHandler,
		fmt.Sprintf("%s/reports", apiPath):           api.GetReportData,
	}

	securedHandlers := make(map[string]http.HandlerFunc)
	for path, handler := range baseHandlers {
		securedHandlers[path] = applyMiddlewareChain(handler, m.APIMiddleware, nil)
	}

	return securedHandlers
}

// GetSecuredStaticHandler returns the static file handler with middleware
func GetSecuredStaticHandler(m *Monigo) http.HandlerFunc {
	return applyMiddlewareChain(serveHtmlSite, m.DashboardMiddleware, m.AuthFunction)
}

func applyMiddlewareChain(handler http.HandlerFunc, middleware []func(http.Handler) http.Handler, authFunc func(*http.Request) bool) http.HandlerFunc {
	var finalHandler http.Handler = http.HandlerFunc(handler)

	if authFunc != nil {
		finalHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	for i := len(middleware) - 1; i >= 0; i-- {
		finalHandler = middleware[i](finalHandler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalHandler.ServeHTTP(w, r)
	})
}

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
		api.ViewFunctionMetrics(w, r)
	case path == fmt.Sprintf("%s/reports", apiPath):
		api.GetReportData(w, r)
	default:
		http.NotFound(w, r)
	}
}

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
		return handleFiberAPI(c, api.ViewFunctionMetrics)
	case path == fmt.Sprintf("%s/reports", apiPath):
		return handleFiberAPI(c, api.GetReportData)
	default:
		c.Status(404).SendString("Not Found")
		return nil
	}
}

func handleFiberAPI(c *fiber.Ctx, handler func(http.ResponseWriter, *http.Request)) error {
	respWriter := &fiberResponseWriter{c: c}
	body := c.Request().Body()

	req, err := http.NewRequest(
		string(c.Request().Header.Method()),
		"http://localhost"+string(c.Request().URI().Path()),
		strings.NewReader(string(body)),
	)
	if err != nil {
		c.Status(500).SendString("Internal Server Error")
		return nil
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	if len(body) > 0 {
		req.ContentLength = int64(len(body))
	}

	handler(respWriter, req)
	return nil
}

// resolveStaticPath maps a URL path to an embedded file path and content type.
func resolveStaticPath(urlPath string) (filePath string, contentType string) {
	baseDir := "static"
	filePath = baseDir + urlPath
	if urlPath == "/" {
		filePath = baseDir + "/index.html"
	} else if urlPath == "/favicon.ico" {
		filePath = baseDir + "/assets/favicon.ico"
	}

	ext := filepath.Ext(filePath)
	ct, ok := staticContentTypes[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	return filePath, ct
}

func serveFiberStaticFiles(c *fiber.Ctx, path string) error {
	filePath, contentType := resolveStaticPath(path)

	file, err := staticFiles.ReadFile(filePath)
	if err != nil {
		c.Status(404).SendString("File not found")
		return nil
	}

	c.Set("Content-Type", contentType)
	return c.Send(file)
}

func serveHtmlSite(w http.ResponseWriter, r *http.Request) {
	filePath, contentType := resolveStaticPath(r.URL.Path)

	file, err := staticFiles.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Could not load "+filePath, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(file)
}

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

// ---- Built-in Security Middleware ----

// BasicAuthMiddleware creates a basic authentication middleware
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
func IPWhitelistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			clientIP := getClientIP(r)
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

// RateLimitMiddleware creates a simple rate limiting middleware.
// The returned stop function should be called during shutdown to release the cleanup goroutine.
func RateLimitMiddleware(requests int, window time.Duration) (mw func(http.Handler) http.Handler, stop func()) {
	type clientInfo struct {
		count     int
		lastReset time.Time
	}

	var mu sync.Mutex
	clients := make(map[string]*clientInfo)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(window * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for ip, info := range clients {
					if time.Since(info.lastReset) > window*2 {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	mw = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()

			mu.Lock()
			client, exists := clients[clientIP]
			if !exists {
				client = &clientInfo{count: 0, lastReset: now}
				clients[clientIP] = client
			}
			if now.Sub(client.lastReset) > window {
				client.count = 0
				client.lastReset = now
			}
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

	stop = cancel
	return
}

// LoggingMiddleware creates a request logging middleware
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)
			logger.Log.Info("request", "method", r.Method, "path", r.URL.Path, "status", wrapped.statusCode, "duration", duration, "remote", r.RemoteAddr)
		})
	}
}

// ---- Helper functions ----

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func isIPAllowed(clientIP, allowedIP string) bool {
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
	return clientIP == allowedIP
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func isStaticFile(path string) bool {
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".map",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	staticPaths := []string{
		"/css/", "/js/", "/assets/", "/images/", "/fonts/", "/static/",
	}

	for _, staticPath := range staticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	if path == "/favicon.ico" {
		return true
	}

	return false
}
