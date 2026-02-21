package core

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/internal/logger"
	"github.com/iyashjayesh/monigo/models"
)

const maxTrackedFunctions = 10000

var (
	functionMetrics = make(map[string]*models.FunctionMetrics)
	basePath        = common.GetBasePath()

	samplingRate atomic.Int64
	callCounters = make(map[string]uint64)
	countersMu   sync.Mutex
)

func init() {
	samplingRate.Store(100)
}

// SetSamplingRate sets the sampling rate for function tracing
func SetSamplingRate(rate int) {
	if rate < 1 {
		rate = 1
	}
	samplingRate.Store(int64(rate))
}

// TraceFunction traces the function and captures the metrics
func TraceFunction(_ context.Context, f func()) {
	name := strings.ReplaceAll(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "/", "-")
	executeFunctionWithProfiling(name, f)
}

// FunctionTraceDetails returns a snapshot copy of the function trace details (thread-safe)
func FunctionTraceDetails() map[string]*models.FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()

	result := make(map[string]*models.FunctionMetrics, len(functionMetrics))
	for k, v := range functionMetrics {
		copied := *v
		result[k] = &copied
	}
	return result
}

// TraceFunctionWithArgs traces a function with parameters and captures the metrics
func TraceFunctionWithArgs(_ context.Context, f interface{}, args ...interface{}) {
	fnValue := reflect.ValueOf(f)
	if fnValue.Kind() != reflect.Func {
		logger.Log.Error("first argument must be a function", "type", fmt.Sprintf("%T", f))
		return
	}

	fnType := fnValue.Type()

	if len(args) != fnType.NumIn() {
		logger.Log.Error("function argument count mismatch", "expected", fnType.NumIn(), "got", len(args))
		return
	}

	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		expectedType := fnType.In(i)

		if !argValue.Type().AssignableTo(expectedType) {
			logger.Log.Error("argument type mismatch", "index", i, "expected", expectedType, "got", argValue.Type())
			return
		}
		argValues[i] = argValue
	}

	name := generateFunctionName(fnValue, fnType)

	executeFunctionWithProfiling(name, func() {
		fnValue.Call(argValues)
	})
}

// TraceFunctionWithReturn traces a function and returns the first result.
func TraceFunctionWithReturn(ctx context.Context, f interface{}, args ...interface{}) interface{} {
	results := TraceFunctionWithReturns(ctx, f, args...)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

// TraceFunctionWithReturns traces a function and returns all results.
func TraceFunctionWithReturns(_ context.Context, f interface{}, args ...interface{}) []interface{} {
	fnValue := reflect.ValueOf(f)
	if fnValue.Kind() != reflect.Func {
		logger.Log.Error("first argument must be a function", "type", fmt.Sprintf("%T", f))
		return nil
	}

	fnType := fnValue.Type()

	if len(args) != fnType.NumIn() {
		logger.Log.Error("function argument count mismatch", "expected", fnType.NumIn(), "got", len(args))
		return nil
	}

	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValue := reflect.ValueOf(arg)
		expectedType := fnType.In(i)

		if !argValue.Type().AssignableTo(expectedType) {
			logger.Log.Error("argument type mismatch", "index", i, "expected", expectedType, "got", argValue.Type())
			return nil
		}
		argValues[i] = argValue
	}

	name := generateFunctionName(fnValue, fnType)

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

func generateFunctionName(fnValue reflect.Value, fnType reflect.Type) string {
	baseName := strings.ReplaceAll(runtime.FuncForPC(fnValue.Pointer()).Name(), "/", "-")

	if fnType.NumIn() > 0 {
		paramTypes := make([]string, fnType.NumIn())
		for i := 0; i < fnType.NumIn(); i++ {
			paramTypes[i] = fnType.In(i).String()
		}
		baseName = fmt.Sprintf("%s(%s)", baseName, strings.Join(paramTypes, ","))
	}

	if fnType.NumOut() > 0 {
		returnTypes := make([]string, fnType.NumOut())
		for i := 0; i < fnType.NumOut(); i++ {
			returnTypes[i] = fnType.Out(i).String()
		}
		baseName = fmt.Sprintf("%s->(%s)", baseName, strings.Join(returnTypes, ","))
	}

	return baseName
}

// sanitizeFileName replaces characters that are invalid in file paths.
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"(", "_", ")", "_",
		"<", "_", ">", "_",
		":", "_", "*", "_",
		"?", "_", "\"", "_",
		"|", "_", " ", "_",
	)
	return replacer.Replace(name)
}

func executeFunctionWithProfiling(name string, fn func()) {
	countersMu.Lock()
	if len(callCounters) > maxTrackedFunctions {
		// Evict oldest entries to prevent unbounded growth.
		for k := range callCounters {
			delete(callCounters, k)
			break
		}
	}
	callCounters[name]++
	count := callCounters[name]
	countersMu.Unlock()

	shouldProfile := count%uint64(samplingRate.Load()) == 0

	initialGoroutines := runtime.NumGoroutine()
	var memStatsBefore runtime.MemStats
	if shouldProfile {
		runtime.ReadMemStats(&memStatsBefore)
	}

	var cpuProfFilePath, memProfFilePath string
	var cpuProfileFile *os.File

	if shouldProfile {
		folderPath := fmt.Sprintf("%s/profiles", basePath)
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			logger.Log.Warn("failed to create profiles directory", "error", err)
		}

		safeName := sanitizeFileName(name)
		cpuProfFilePath = filepath.Join(folderPath, fmt.Sprintf("%s_cpu.prof", safeName))
		memProfFilePath = filepath.Join(folderPath, fmt.Sprintf("%s_mem.prof", safeName))

		var err error
		cpuProfileFile, err = StartCPUProfile(cpuProfFilePath)
		if err != nil {
			logger.Log.Warn("failed to start CPU profile", "error", err)
		}
	}

	start := time.Now()
	fn()
	elapsed := time.Since(start)

	if shouldProfile {
		StopCPUProfile(cpuProfileFile)
		if err := WriteHeapProfile(memProfFilePath); err != nil {
			logger.Log.Warn("failed to write heap profile", "error", err)
		}
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

	if len(functionMetrics) > maxTrackedFunctions {
		// Evict one arbitrary entry to cap memory.
		for k := range functionMetrics {
			delete(functionMetrics, k)
			break
		}
	}

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
	_, err := exec.LookPath("go")
	if err != nil {
		logger.Log.Warn("'go' command not found in PATH, pprof reports will be unavailable")
		return models.FunctionTraceDetails{
			FunctionName: name,
			CoreProfile: models.Profiles{
				CPU: "Error: 'go' command not found. pprof reports require the Go SDK.",
				Mem: "Error: 'go' command not found. pprof reports require the Go SDK.",
			},
			FunctionCodeTrace: "Error: 'go' command not found.",
		}
	}

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

	return models.FunctionTraceDetails{
		FunctionName: name,
		CoreProfile: models.Profiles{
			CPU: executePprof(metrics.CPUProfileFilePath, reportType),
			Mem: executePprof(metrics.MemProfileFilePath, reportType),
		},
		FunctionCodeTrace: codeStack,
	}
}
