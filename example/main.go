package main

import (
	"log"
	"math"
	"net/http"

	"github.com/iyashjayesh/monigo"
)

func main() {
	monigoInstance := monigo.NewBuilder().
		WithServiceName("data-api").
		WithPort(8080).
		WithRetentionPeriod("4d").
		WithDataPointsSyncFrequency("5m").
		Build()

	go func() {
		if err := monigoInstance.Start(); err != nil {
			log.Fatalf("Failed to start MoniGo: %v", err)
		}
	}()
	log.Printf("Monigo dashboard started at port %d\n", monigoInstance.GetRunningPort())

	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/api2", apiHandler2)
	log.Println("Your application started at port 8000")
	http.ListenAndServe(":8000", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(r.Context(), highMemoryUsage)
	w.Write([]byte("API1 response memexpensiveFunc"))
}

func apiHandler2(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(r.Context(), highCPUUsage)
	w.Write([]byte("API2 response cpuexpensiveFunc"))
}

func highMemoryUsage() {
	largeSlice := make([]float64, 1e8)
	for i := 0; i < len(largeSlice); i++ {
		largeSlice[i] = float64(i)
	}
}

func highCPUUsage() {
	var sum float64
	for i := 0; i < 1e8; i++ {
		sum += math.Sqrt(float64(i))
	}
	_ = sum
}
