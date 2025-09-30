package main

import (
	"log"
	"net/http"

	"github.com/iyashjayesh/monigo"
)

func main() {
	// Initialize MoniGo without starting the dashboard
	monigoInstance := &monigo.Monigo{
		ServiceName:             "standard-mux-integration-example",
		DataPointsSyncFrequency: "5m",
		DataRetentionPeriod:     "7d",
		TimeZone:                "Local",
		CustomBaseAPIPath:       "/monigo/api/v1", // Custom API path
	}

	// Initialize MoniGo (this sets up metrics collection but doesn't start the dashboard)
	monigoInstance.Initialize()

	// Create standard HTTP mux
	mux := http.NewServeMux()

	// Add your own routes first (these take priority)
	mux.HandleFunc("/api/users", usersHandler)
	mux.HandleFunc("/api/orders", ordersHandler)
	mux.HandleFunc("/health", healthHandler)

	// Get MoniGo unified handler that handles both API and static files
	unifiedHandler := monigo.GetUnifiedHandler("/monigo/api/v1")
	mux.HandleFunc("/", unifiedHandler)

	log.Println("Server starting on :8080")
	log.Println("MoniGo dashboard available at: http://localhost:8080/")
	log.Println("MoniGo API available at: http://localhost:8080/monigo/api/v1/")
	log.Println("Your API available at: http://localhost:8080/api/")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate some work
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Users endpoint", "count": 42}`))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	// Trace this function for monitoring
	monigo.TraceFunction(func() {
		// Simulate CPU intensive work
		for i := 0; i < 1000000; i++ {
			_ = i * i
		}
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Orders endpoint", "count": 15}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "standard-mux-integration-example"}`))
}
