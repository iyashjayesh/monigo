package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/iyashjayesh/monigo"
)

type User struct {
	ID    string
	Name  string
	Email string
}

type Item struct {
	Name  string
	Price float64
}

type Result struct {
	Success bool
	Message string
	Data    interface{}
}

func main() {
	monigoInstance := monigo.NewBuilder().
		WithServiceName("comprehensive-trace-api").
		WithPort(8080).
		WithDataPointsSyncFrequency("5s").
		WithRetentionPeriod("4d").
		WithTimeZone("Local").
		Build()

	go monigoInstance.Start()
	log.Println("Monigo dashboard started at port 8080")

	http.HandleFunc("/api/legacy", legacyHandler)
	http.HandleFunc("/api/user", userHandler)
	http.HandleFunc("/api/calculate", calculateHandler)
	http.HandleFunc("/api/process", processHandler)
	http.HandleFunc("/api/validate", validateHandler)
	http.HandleFunc("/api/memory", memoryHandler)
	http.HandleFunc("/api/cpu", cpuHandler)

	log.Println("Comprehensive trace example started at port 8000")
	log.Println("Visit http://localhost:8080 to see the MoniGo dashboard")
	http.ListenAndServe(":8000", nil)
}

func legacyHandler(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(r.Context(), func() {
		time.Sleep(100 * time.Millisecond)
		_ = make([]byte, 1024*1024)
	})
	w.Write([]byte("Legacy tracing method still works!"))
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	userName := r.URL.Query().Get("name")
	if userID == "" {
		userID = "default"
	}
	if userName == "" {
		userName = "Anonymous"
	}

	monigo.TraceFunctionWithArgs(r.Context(), processUser, userID, userName)
	user := monigo.TraceFunctionWithReturn(r.Context(), createUser, userID, userName).(User)

	w.Write([]byte(fmt.Sprintf("Processed user: %+v", user)))
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	items := []Item{
		{Name: "Laptop", Price: 999.99},
		{Name: "Mouse", Price: 29.99},
		{Name: "Keyboard", Price: 79.99},
	}

	total := monigo.TraceFunctionWithReturn(r.Context(), calculateTotal, items).(float64)

	discount := 0.1
	finalTotal := monigo.TraceFunctionWithReturn(r.Context(), applyDiscount, total, discount).(float64)

	w.Write([]byte(fmt.Sprintf("Total: $%.2f, Final Total: $%.2f", total, finalTotal)))
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		data = "default-data"
	}

	results := monigo.TraceFunctionWithReturns(r.Context(), processData, data)
	if len(results) >= 2 {
		result := results[0].(Result)
		err := results[1].(error)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}
		w.Write([]byte(fmt.Sprintf("Processed: %+v", result)))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unexpected result format"))
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	valueStr := r.URL.Query().Get("value")
	value := 0
	if valueStr != "" {
		if parsed, err := strconv.Atoi(valueStr); err == nil {
			value = parsed
		}
	}

	results := monigo.TraceFunctionWithReturns(r.Context(), validateValue, value)
	if len(results) >= 3 {
		valid := results[0].(bool)
		message := results[1].(string)
		err := results[2].(error)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}
		w.Write([]byte(fmt.Sprintf("Valid: %t, Message: %s", valid, message)))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unexpected result format"))
	}
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(r.Context(), highMemoryUsage)
	w.Write([]byte("Memory-intensive function traced!"))
}

func cpuHandler(w http.ResponseWriter, r *http.Request) {
	monigo.TraceFunction(r.Context(), highCPUUsage)
	w.Write([]byte("CPU-intensive function traced!"))
}

func processUser(userID, userName string) {
	time.Sleep(50 * time.Millisecond)
	_ = make([]byte, 512*1024)
	for i := 0; i < 100000; i++ {
		_ = math.Sqrt(float64(i))
	}
}

func createUser(userID, userName string) User {
	time.Sleep(30 * time.Millisecond)
	return User{
		ID:    userID,
		Name:  userName,
		Email: fmt.Sprintf("%s@example.com", userID),
	}
}

func calculateTotal(items []Item) float64 {
	time.Sleep(20 * time.Millisecond)
	var total float64
	for _, item := range items {
		total += item.Price
	}
	for i := 0; i < 50000; i++ {
		_ = math.Sqrt(total)
	}
	return total
}

func applyDiscount(total, discount float64) float64 {
	time.Sleep(10 * time.Millisecond)
	return total * (1 - discount)
}

func processData(data string) (Result, error) {
	time.Sleep(30 * time.Millisecond)
	if data == "error" {
		return Result{}, fmt.Errorf("processing error")
	}
	for i := 0; i < 50000; i++ {
		_ = len(data) * i
	}
	return Result{
		Success: true,
		Message: "Data processed successfully",
		Data:    fmt.Sprintf("Processed: %s", data),
	}, nil
}

func validateValue(value int) (bool, string, error) {
	time.Sleep(10 * time.Millisecond)
	if value < 0 {
		return false, "Value cannot be negative", nil
	}
	if value > 100 {
		return false, "Value cannot be greater than 100", nil
	}
	for i := 0; i < value*1000; i++ {
		_ = i * i
	}
	return true, fmt.Sprintf("Value %d is valid", value), nil
}

func highMemoryUsage() {
	largeSlice := make([]float64, 1e7)
	for i := 0; i < len(largeSlice); i++ {
		largeSlice[i] = float64(i)
	}
}

func highCPUUsage() {
	var sum float64
	for i := 0; i < 1e7; i++ {
		sum += math.Sqrt(float64(i))
	}
	_ = sum
}

