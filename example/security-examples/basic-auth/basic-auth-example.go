package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/iyashjayesh/monigo"
)

func main() {
	// Rate limit middleware returns (mw, stop) - call stop on shutdown
	rateLimitMW, stop := monigo.RateLimitMiddleware(100, time.Minute)
	defer stop()

	// Initialize MoniGo with basic authentication
	monigoInstance := &monigo.Monigo{
		ServiceName:             "basic-auth-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1",

		// Security Configuration
		DashboardMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "password123"),
			monigo.LoggingMiddleware(),
		},
		APIMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "password123"),
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
	log.Println("Username: admin, Password: password123")
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
