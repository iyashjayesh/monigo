package core

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
)

var functionTraceDetails = make(map[string]string)

// Metrics struct to hold memory and CPU statistics
type Metrics struct {
	Alloc      uint64 // Memory allocated (bytes)
	TotalAlloc uint64 // Total memory allocated
	Sys        uint64 // Memory obtained from the system
	NumGC      uint32 // Number of garbage collections
	Goroutines int    // Number of goroutines
	NumCPU     int    // Number of CPUs
}

// captureMetrics captures memory and CPU statistics
func captureMetrics() Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Metrics{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
		Goroutines: runtime.NumGoroutine(),
		NumCPU:     runtime.NumCPU(),
	}
}

// trace logs the stack trace, execution time, and system metrics
func trace(start time.Time, beforeMetrics Metrics, fnName string, depth int) string {
	// Capture stack trace
	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Function call stack trace for %s:\n", fnName))
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	// Calculate function execution time
	elapsed := time.Since(start)
	sb.WriteString(fmt.Sprintf("Execution time: %s\n", elapsed))

	// Log system metrics (after function execution)
	afterMetrics := captureMetrics()
	sb.WriteString("System Metrics Before Execution:\n")
	sb.WriteString(fmt.Sprintf("Alloc = %v MiB\n", beforeMetrics.Alloc/1024/1024))
	sb.WriteString(fmt.Sprintf("TotalAlloc = %v MiB\n", beforeMetrics.TotalAlloc/1024/1024))
	sb.WriteString(fmt.Sprintf("Sys = %v MiB\n", beforeMetrics.Sys/1024/1024))
	sb.WriteString(fmt.Sprintf("NumGC = %v\n", beforeMetrics.NumGC))
	sb.WriteString(fmt.Sprintf("Goroutines = %v\n", beforeMetrics.Goroutines))
	sb.WriteString(fmt.Sprintf("CPUs = %v\n", beforeMetrics.NumCPU))

	sb.WriteString("System Metrics After Execution:\n")
	sb.WriteString(fmt.Sprintf("Alloc = %v MiB\n", afterMetrics.Alloc/1024/1024))
	sb.WriteString(fmt.Sprintf("TotalAlloc = %v MiB\n", afterMetrics.TotalAlloc/1024/1024))
	sb.WriteString(fmt.Sprintf("Sys = %v MiB\n", afterMetrics.Sys/1024/1024))
	sb.WriteString(fmt.Sprintf("NumGC = %v\n", afterMetrics.NumGC))
	sb.WriteString(fmt.Sprintf("Goroutines = %v\n", afterMetrics.Goroutines))
	sb.WriteString(fmt.Sprintf("CPUs = %v\n", afterMetrics.NumCPU))

	return sb.String()
}

// logFunctionExecution logs the performance metrics for the given function
func TraceFunctionExecution(fn func(), fnName string) {
	// Initial stack depth
	initialDepth := 10
	maxDepth := 100
	increment := 10
	depth := initialDepth

	for depth <= maxDepth {
		beforeMetrics := captureMetrics()
		start := time.Now()
		defer func() {
			if _, ok := functionTraceDetails[fnName]; ok {
				functionTraceDetails[fnName] = functionTraceDetails[fnName] + trace(start, beforeMetrics, fnName, depth)
			} else {
				functionTraceDetails[fnName] = trace(start, beforeMetrics, fnName, depth)
			}
		}()

		fn()

		if isSufficientDepth(depth) { // Checking if the depth is sufficient based on analysis
			break
		}

		depth += increment // Increasing the depth for the next iteration
	}
}

// isSufficientDepth analyzes if the current stack depth is sufficient
// For simplicity, we assume a fixed threshold here
func isSufficientDepth(depth int) bool {
	// In a real scenario, you might analyze the depth or use some heuristics
	return depth >= 20 // Example threshold
}

func TraceFunction(f func()) {

	if f == nil {
		log.Println("Function is nil, cannot register")
		return
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	log.Println("Tracing function:", funcName)
	TraceFunctionExecution(f, funcName)
}

func FunctionTraceDetails() map[string]string {
	return functionTraceDetails
}
