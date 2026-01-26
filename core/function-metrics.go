package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

var (
	functionMetrics = make(map[string]*models.FunctionMetrics)
	basePath        = common.GetBasePath()

	// Sampling configuration
	samplingRate = 100 // Trace 1 in 100 calls by default
	callCounters = make(map[string]uint64)
	countersMu   sync.Mutex
)

// SetSamplingRate sets the sampling rate for function tracing
func SetSamplingRate(rate int) {
	if rate < 1 {
		rate = 1
	}
	samplingRate = rate
}

// TraceFunction traces the function and captures the metrics
// This is the original function maintained for backward compatibility
func TraceFunction(f func()) {
	name := strings.ReplaceAll(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "/", "-") // Getting the name of the function
	executeFunctionWithProfiling(name, f)
}

// FunctionTraceDetails returns the function trace details
func FunctionTraceDetails() map[string]*models.FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()

	return functionMetrics
}

// TraceFunctionWithArgs traces a function with parameters and captures the metrics
// This function uses reflection to call functions with arbitrary signatures
func TraceFunctionWithArgs(f interface{}, args ...interface{}) {
	// Validate that f is a function
	fnValue := reflect.ValueOf(f)
	if fnValue.Kind() != reflect.Func {
		log.Printf("[MoniGo] Error: first argument must be a function, got %T", f)
		return
	}

	// Get function type information
	fnType := fnValue.Type()

	// Validate argument count
	if len(args) != fnType.NumIn() {
		log.Printf("[MoniGo] Error: function expects %d arguments, got %d", fnType.NumIn(), len(args))
		return
	}

	// Convert arguments to reflect.Values and validate types
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		expectedType := fnType.In(i)

		// Check if types are compatible
		if !argValue.Type().AssignableTo(expectedType) {
			log.Printf("[MoniGo] Error: argument %d type mismatch. Expected %v, got %v", i, expectedType, argValue.Type())
			return
		}
		argValues[i] = argValue
	}

	// Generate function name with parameter types for better identification
	name := generateFunctionName(fnValue, fnType)

	// Execute the function with profiling
	executeFunctionWithProfiling(name, func() {
		fnValue.Call(argValues)
	})
}

// TraceFunctionWithReturn traces a function with parameters and return values
// Returns the first result of the function call (for backward compatibility)
func TraceFunctionWithReturn(f interface{}, args ...interface{}) interface{} {
	results := TraceFunctionWithReturns(f, args...)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// TraceFunctionWithReturns traces a function with parameters and return values
// Returns all results of the function call as a slice of interface{}
func TraceFunctionWithReturns(f interface{}, args ...interface{}) []interface{} {
	// Validate that f is a function
	fnValue := reflect.ValueOf(f)
	if fnValue.Kind() != reflect.Func {
		log.Printf("[MoniGo] Error: first argument must be a function, got %T", f)
		return nil
	}

	// Get function type information
	fnType := fnValue.Type()

	// Validate argument count
	if len(args) != fnType.NumIn() {
		log.Printf("[MoniGo] Error: function expects %d arguments, got %d", fnType.NumIn(), len(args))
		return nil
	}

	// Convert arguments to reflect.Values and validate types
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		expectedType := fnType.In(i)

		// Check if types are compatible
		if !argValue.Type().AssignableTo(expectedType) {
			log.Printf("[MoniGo] Error: argument %d type mismatch. Expected %v, got %v", i, expectedType, argValue.Type())
			return nil
		}
		argValues[i] = argValue
	}

	// Generate function name with parameter types for better identification
	name := generateFunctionName(fnValue, fnType)

	// Execute the function with profiling and capture return values
	var results []interface{}
	executeFunctionWithProfiling(name, func() {
		reflectResults := fnValue.Call(argValues)
		results = make([]interface{}, len(reflectResults))
		for i, result := range reflectResults {
			results[i] = result.Interface()
		}
	})

	return results
}

// generateFunctionName creates a descriptive name for the function including parameter types
func generateFunctionName(fnValue reflect.Value, fnType reflect.Type) string {
	// Get the base function name
	baseName := strings.ReplaceAll(runtime.FuncForPC(fnValue.Pointer()).Name(), "/", "-")

	// Add parameter type information for better identification
	if fnType.NumIn() > 0 {
		var paramTypes []string
		for i := 0; i < fnType.NumIn(); i++ {
			paramTypes = append(paramTypes, fnType.In(i).String())
		}
		baseName = fmt.Sprintf("%s(%s)", baseName, strings.Join(paramTypes, ","))
	}

	// Add return type information if there are return values
	if fnType.NumOut() > 0 {
		var returnTypes []string
		for i := 0; i < fnType.NumOut(); i++ {
			returnTypes = append(returnTypes, fnType.Out(i).String())
		}
		baseName = fmt.Sprintf("%s->(%s)", baseName, strings.Join(returnTypes, ","))
	}

	return baseName
}

// executeFunctionWithProfiling contains the common profiling logic with sampling
func executeFunctionWithProfiling(name string, fn func()) {
	countersMu.Lock()
	callCounters[name]++
	count := callCounters[name]
	countersMu.Unlock()

	shouldProfile := count%uint64(samplingRate) == 0

	initialGoroutines := runtime.NumGoroutine()
	var memStatsBefore runtime.MemStats
	if shouldProfile {
		runtime.ReadMemStats(&memStatsBefore)
	}

	var cpuProfFilePath, memProfFilePath string
	var cpuProfileFile *os.File

	if shouldProfile {
		folderPath := fmt.Sprintf("%s/profiles", basePath)
		_ = os.MkdirAll(folderPath, os.ModePerm)

		cpuProfFilePath = filepath.Join(folderPath, fmt.Sprintf("%s_cpu.prof", name))
		memProfFilePath = filepath.Join(folderPath, fmt.Sprintf("%s_mem.prof", name))

		var err error
		cpuProfileFile, err = StartCPUProfile(cpuProfFilePath)
		if err != nil {
			log.Printf("[MoniGo] Warning: failed to start CPU profile: %v", err)
		}
	}

	start := time.Now()
	fn()
	elapsed := time.Since(start)

	if shouldProfile {
		StopCPUProfile(cpuProfileFile)
		_ = WriteHeapProfile(memProfFilePath)
	}

	finalGoroutines := runtime.NumGoroutine() - initialGoroutines
	if finalGoroutines < 0 {
		finalGoroutines = 0
	}

	var memoryUsage uint64
	if shouldProfile {
		var memStatsAfter runtime.MemStats
		runtime.ReadMemStats(&memStatsAfter)
		if memStatsAfter.Alloc >= memStatsBefore.Alloc {
			memoryUsage = memStatsAfter.Alloc - memStatsBefore.Alloc
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// Update or create metrics
	if m, exists := functionMetrics[name]; exists {
		m.FunctionLastRanAt = start
		m.ExecutionTime = elapsed
		m.GoroutineCount = finalGoroutines
		if shouldProfile {
			m.MemoryUsage = memoryUsage
			m.CPUProfileFilePath = cpuProfFilePath
			m.MemProfileFilePath = memProfFilePath
		}
	} else {
		functionMetrics[name] = &models.FunctionMetrics{
			FunctionLastRanAt:  start,
			ExecutionTime:      elapsed,
			GoroutineCount:     finalGoroutines,
			MemoryUsage:        memoryUsage,
			CPUProfileFilePath: cpuProfFilePath,
			MemProfileFilePath: memProfFilePath,
		}
	}
}

// ViewFunctionMetrics generates the function metrics
func ViewFunctionMetrics(name, reportType string, metrics *models.FunctionMetrics) models.FunctionTraceDetails {
	// Check if 'go' command is available
	_, err := exec.LookPath("go")
	if err != nil {
		log.Printf("[MoniGo] Warning: 'go' command not found in PATH. pprof reports will be unavailable.")
		return models.FunctionTraceDetails{
			FunctionName: name,
			CoreProfile: models.Profiles{
				CPU: "Error: 'go' command not found. pprof reports require the Go SDK.",
				Mem: "Error: 'go' command not found. pprof reports require the Go SDK.",
			},
			FunctionCodeTrace: "Error: 'go' command not found.",
		}
	}

	// Function to execute the pprof command and return the output or log an error
	executePprof := func(profileFilePath, reportType string) string {
		if profileFilePath == "" {
			return "Error: Profile file path is empty"
		}
		cmd := exec.Command("go", "tool", "pprof", "-"+reportType, profileFilePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error executing pprof: %v\nOutput: %s", err, string(output))
		}
		return string(output)
	}

	// Generating the function code stack trace for CPU profile
	var codeStack string
	if metrics.CPUProfileFilePath != "" {
		codeStackView := exec.Command("go", "tool", "pprof", "-list", name, metrics.CPUProfileFilePath)
		output, err := codeStackView.CombinedOutput()
		if err != nil {
			codeStack = fmt.Sprintf("Error generating code trace: %v\nOutput: %s", err, string(output))
		} else {
			codeStack = string(output)
		}
	}

	// Return the function trace details
	return models.FunctionTraceDetails{
		FunctionName: name,
		CoreProfile: models.Profiles{
			CPU: executePprof(metrics.CPUProfileFilePath, reportType),
			Mem: executePprof(metrics.MemProfileFilePath, reportType),
		},
		FunctionCodeTrace: codeStack,
	}
}
