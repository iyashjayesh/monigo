package main

import (
	"log"
	"math"
	"net/http"

	"github.com/iyashjayesh/monigo"
)

func main() {

	// New way: Use Builder Pattern
	monigoInstance := monigo.NewBuilder().
		WithServiceName("data-api").
		WithPort(8080).
		WithRetentionPeriod("4d").
		WithDataPointsSyncFrequency("5m").
		Build()

	// Traditional way: Start MoniGo dashboard on a separate port
	go func() {
		if err := monigoInstance.Start(); err != nil {
			log.Fatalf("Failed to start MoniGo: %v", err)
		}
	}()
	log.Println("Monigo dashboard started at port 8080")

	// Your application runs on a different port
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/api2", apiHandler2)
	log.Println("Your application started at port 8000")
	http.ListenAndServe(":8000", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(highMemoryUsage) // Trace function, when the function is called, it will be traced
	w.Write([]byte("API1 response memexpensiveFunc"))
}

func apiHandler2(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(highCPUUsage) // Trace function, when the function is called, it will be traced
	w.Write([]byte("API2 response cpuexpensiveFunc"))
}

func highMemoryUsage() {
	// Simulate high memory usage by allocating a large slice
	largeSlice := make([]float64, 1e8) // 100 million elements
	for i := 0; i < len(largeSlice); i++ {
		largeSlice[i] = float64(i)
	}
}

func highCPUUsage() {
	// Simulate high CPU usage by performing heavy computations
	var sum float64
	for i := 0; i < 1e8; i++ { // 100 million iterations
		sum += math.Sqrt(float64(i))
	}
}
