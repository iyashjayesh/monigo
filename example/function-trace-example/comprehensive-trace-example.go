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

// User represents a user in our system
type User struct {
	ID    string
	Name  string
	Email string
}

// Item represents an item for calculation
type Item struct {
	Name  string
	Price float64
}

// Result represents a processing result
type Result struct {
	Success bool
	Message string
	Data    interface{}
}

func main() {
	// Initialize MoniGo
	// Initialize MoniGo
	monigoInstance := monigo.NewBuilder().
		WithServiceName("comprehensive-trace-api").
		WithPort(8080).
		WithDataPointsSyncFrequency("5s").
		WithRetentionPeriod("4d").
		WithTimeZone("Local").
		Build()

	// Start MoniGo dashboard
	go monigoInstance.Start()
	log.Println("Monigo dashboard started at port 8080")

	// Set up HTTP handlers demonstrating all tracing scenarios
	http.HandleFunc("/api/legacy", legacyHandler)       // Original TraceFunction
	http.HandleFunc("/api/user", userHandler)           // TraceFunctionWithArgs
	http.HandleFunc("/api/calculate", calculateHandler) // TraceFunctionWithReturn
	http.HandleFunc("/api/process", processHandler)     // TraceFunctionWithReturns (2 returns)
	http.HandleFunc("/api/validate", validateHandler)   // TraceFunctionWithReturns (3 returns)
	http.HandleFunc("/api/memory", memoryHandler)       // Memory-intensive function
	http.HandleFunc("/api/cpu", cpuHandler)             // CPU-intensive function

	log.Println("Comprehensive trace example started at port 8000")
	log.Println("Visit http://localhost:8080 to see the MoniGo dashboard")
	log.Println("Test endpoints:")
	log.Println("  GET http://localhost:8000/api/legacy")
	log.Println("  GET http://localhost:8000/api/user?id=123&name=John")
	log.Println("  GET http://localhost:8000/api/calculate?count=5")
	log.Println("  GET http://localhost:8000/api/process?data=test")
	log.Println("  GET http://localhost:8000/api/validate?value=42")
	log.Println("  GET http://localhost:8000/api/memory")
	log.Println("  GET http://localhost:8000/api/cpu")

	http.ListenAndServe(":8000", nil)
}

// ============================================================================
// HANDLERS DEMONSTRATING DIFFERENT TRACING METHODS
// ============================================================================

// legacyHandler demonstrates backward compatibility with the original TraceFunction
func legacyHandler(w http.ResponseWriter, r *http.Request) {
	// OLD WAY: Still works for backward compatibility
	monigo.TraceFunction(func() {
		// Simulate some work
		time.Sleep(100 * time.Millisecond)
		_ = make([]byte, 1024*1024) // 1MB allocation
	})

	w.Write([]byte("Legacy tracing method still works!"))
}

// userHandler demonstrates TraceFunctionWithArgs - functions with parameters
func userHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	userName := r.URL.Query().Get("name")

	if userID == "" {
		userID = "default"
	}
	if userName == "" {
		userName = "Anonymous"
	}

	// NEW WAY: Direct function tracing with parameters
	monigo.TraceFunctionWithArgs(processUser, userID, userName)

	// Alternative: You can also trace functions that return values
	user := monigo.TraceFunctionWithReturn(createUser, userID, userName).(User)

	w.Write([]byte(fmt.Sprintf("Processed user: %+v", user)))
}

// calculateHandler demonstrates TraceFunctionWithReturn - functions with single return
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate some items for calculation
	items := []Item{
		{Name: "Laptop", Price: 999.99},
		{Name: "Mouse", Price: 29.99},
		{Name: "Keyboard", Price: 79.99},
	}

	// NEW WAY: Trace function with return value
	total := monigo.TraceFunctionWithReturn(calculateTotal, items).(float64)

	// Also demonstrate tracing a function with multiple parameters
	discount := 0.1
	finalTotal := monigo.TraceFunctionWithReturn(applyDiscount, total, discount).(float64)

	w.Write([]byte(fmt.Sprintf("Total: $%.2f, Final Total: $%.2f", total, finalTotal)))
}

// processHandler demonstrates TraceFunctionWithReturns - functions with 2 returns (result, error)
func processHandler(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		data = "default-data"
	}

	// NEW WAY: Trace function with multiple returns
	results := monigo.TraceFunctionWithReturns(processData, data)

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

// validateHandler demonstrates TraceFunctionWithReturns - functions with 3 returns
func validateHandler(w http.ResponseWriter, r *http.Request) {
	valueStr := r.URL.Query().Get("value")
	value := 0 // default
	if valueStr != "" {
		if parsed, err := strconv.Atoi(valueStr); err == nil {
			value = parsed
		}
	}

	// NEW WAY: Trace function with multiple returns
	results := monigo.TraceFunctionWithReturns(validateValue, value)

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

// memoryHandler demonstrates memory-intensive function tracing
func memoryHandler(w http.ResponseWriter, r *http.Request) {
	// NEW WAY: Trace memory-intensive function
	monigo.TraceFunctionWithArgs(highMemoryUsage)

	w.Write([]byte("Memory-intensive function traced!"))
}

// cpuHandler demonstrates CPU-intensive function tracing
func cpuHandler(w http.ResponseWriter, r *http.Request) {
	// NEW WAY: Trace CPU-intensive function
	monigo.TraceFunctionWithArgs(highCPUUsage)

	w.Write([]byte("CPU-intensive function traced!"))
}

// ============================================================================
// BUSINESS LOGIC FUNCTIONS WITH DIFFERENT SIGNATURES
// ============================================================================

// processUser processes a user with string parameters (no return)
func processUser(userID, userName string) {
	// Simulate some processing work
	time.Sleep(50 * time.Millisecond)

	// Simulate memory allocation
	_ = make([]byte, 512*1024) // 512KB allocation

	// Simulate some CPU work
	for i := 0; i < 100000; i++ {
		_ = math.Sqrt(float64(i))
	}
}

// createUser creates a user and returns it (single return)
func createUser(userID, userName string) User {
	// Simulate some processing
	time.Sleep(30 * time.Millisecond)

	return User{
		ID:    userID,
		Name:  userName,
		Email: fmt.Sprintf("%s@example.com", userID),
	}
}

// calculateTotal calculates the total price of items (single return)
func calculateTotal(items []Item) float64 {
	// Simulate some processing
	time.Sleep(20 * time.Millisecond)

	var total float64
	for _, item := range items {
		total += item.Price
	}

	// Simulate some CPU work
	for i := 0; i < 50000; i++ {
		_ = math.Sqrt(total)
	}

	return total
}

// applyDiscount applies a discount to a total (single return)
func applyDiscount(total, discount float64) float64 {
	// Simulate some processing
	time.Sleep(10 * time.Millisecond)

	return total * (1 - discount)
}

// processData processes data and returns (Result, error) - 2 returns
func processData(data string) (Result, error) {
	// Simulate some processing
	time.Sleep(30 * time.Millisecond)

	// Simulate validation
	if data == "error" {
		return Result{}, fmt.Errorf("processing error")
	}

	// Simulate CPU work
	for i := 0; i < 50000; i++ {
		_ = len(data) * i
	}

	return Result{
		Success: true,
		Message: "Data processed successfully",
		Data:    fmt.Sprintf("Processed: %s", data),
	}, nil
}

// validateValue validates a value and returns (bool, string, error) - 3 returns
func validateValue(value int) (bool, string, error) {
	// Simulate some processing
	time.Sleep(10 * time.Millisecond)

	// Simulate validation logic
	if value < 0 {
		return false, "Value cannot be negative", nil
	}

	if value > 100 {
		return false, "Value cannot be greater than 100", nil
	}

	// Simulate some CPU work
	for i := 0; i < value*1000; i++ {
		_ = i * i
	}

	return true, fmt.Sprintf("Value %d is valid", value), nil
}

// highMemoryUsage simulates high memory usage (no parameters, no return)
func highMemoryUsage() {
	// Simulate high memory usage by allocating a large slice
	largeSlice := make([]float64, 1e7) // 10 million elements
	for i := 0; i < len(largeSlice); i++ {
		largeSlice[i] = float64(i)
	}
}

// highCPUUsage simulates high CPU usage (no parameters, no return)
func highCPUUsage() {
	// Simulate high CPU usage by performing heavy computations
	var sum float64
	for i := 0; i < 1e7; i++ { // 10 million iterations
		sum += math.Sqrt(float64(i))
	}
}
