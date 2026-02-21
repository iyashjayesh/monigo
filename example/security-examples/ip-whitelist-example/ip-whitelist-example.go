package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/iyashjayesh/monigo"
)

func main() {
	// Rate limit middleware returns (mw, stop) - call stop on shutdown
	rateLimitMW, stop := monigo.RateLimitMiddleware(150, time.Minute)
	defer stop()

	// Initialize MoniGo with IP whitelist
	monigoInstance := &monigo.Monigo{
		ServiceName:             "ip-whitelist-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1",

		// Security Configuration - simplified for testing
		DashboardMiddleware: []func(http.Handler) http.Handler{
			allowLocalhostMiddleware(),
			monigo.LoggingMiddleware(),
		},
		APIMiddleware: []func(http.Handler) http.Handler{
			allowLocalhostMiddleware(),
			rateLimitMW,
		},
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create standard HTTP mux
	mux := http.NewServeMux()

	// Add your own routes first (these take priority)
	mux.HandleFunc("/api/users", usersHandler)
	mux.HandleFunc("/api/orders", ordersHandler)
	mux.HandleFunc("/health", healthHandler)

	// Register MoniGo with security middleware
	monigo.RegisterSecuredDashboardHandlers(mux, monigoInstance, "/monigo/api/v1")

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("Access restricted to: 127.0.0.1, 192.168.1.0/24, 10.0.0.0/8")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Trace this function for monitoring
	monigo.TraceFunction(context.Background(), func() {
		// Simulate some work
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Users endpoint", "count": 42}`))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	// Trace this function for monitoring
	monigo.TraceFunction(context.Background(), func() {
		// Simulate some work
		_ = make([]byte, 512*1024) // 512KB allocation
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Orders endpoint", "count": 15}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// allowLocalhostMiddleware allows only localhost connections
func allowLocalhostMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip IP check for static files
			if isStaticFile(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP
			clientIP := getClientIP(r)
			log.Printf("[IP WHITELIST] Request from IP: %s, RemoteAddr: %s, Path: %s", clientIP, r.RemoteAddr, r.URL.Path)

			// Allow localhost in various formats
			allowedIPs := []string{
				"127.0.0.1",
				"::1",
				"localhost",
			}

			// Check if IP is allowed
			for _, allowedIP := range allowedIPs {
				if clientIP == allowedIP {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check if it's a localhost connection by parsing RemoteAddr
			if strings.Contains(r.RemoteAddr, "127.0.0.1") || strings.Contains(r.RemoteAddr, "::1") {
				next.ServeHTTP(w, r)
				return
			}

			log.Printf("[IP WHITELIST] BLOCKED: IP %s not in whitelist", clientIP)
			http.Error(w, "Forbidden - IP not whitelisted", http.StatusForbidden)
		})
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
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

// isStaticFile checks if the request path is for a static file
func isStaticFile(path string) bool {
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".map", ".json", ".xml",
		".pdf", ".zip", ".txt", ".html", ".htm",
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
