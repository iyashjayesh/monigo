package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iyashjayesh/monigo"
)

func main() {
	// Rate limit middleware returns (mw, stop) - call stop on shutdown
	rateLimitMW, stop := monigo.RateLimitMiddleware(250, time.Minute)
	defer stop()

	// Initialize MoniGo with security middleware
	monigoInstance := &monigo.Monigo{
		ServiceName:             "chi-secured-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1",

		// Security Configuration
		DashboardMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "chi-secure-2024"),
			monigo.LoggingMiddleware(),
		},
		APIMiddleware: []func(http.Handler) http.Handler{
			monigo.BasicAuthMiddleware("admin", "chi-secure-2024"),
			rateLimitMW,
		},
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create Chi router
	r := chi.NewRouter()

	// Add Chi's built-in middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Add your own routes first (these take priority)
	r.Get("/api/users", usersHandler)
	r.Post("/api/orders", ordersHandler)
	r.Get("/health", healthHandler)

	// Get MoniGo secured unified handler
	unifiedHandler := monigo.GetSecuredUnifiedHandler(monigoInstance, "/monigo/api/v1")

	// Register MoniGo with Chi - Chi works directly with http.Handler
	r.Mount("/", http.HandlerFunc(unifiedHandler))

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("Username: admin, Password: chi-secure-2024")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := http.ListenAndServe(":8080", r); err != nil {
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
