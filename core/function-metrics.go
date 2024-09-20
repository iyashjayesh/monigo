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
	"time"

	"github.com/iyashjayesh/monigo/common"
	"github.com/iyashjayesh/monigo/models"
)

var (
	functionMetrics = make(map[string]*models.FunctionMetrics)
	basePath        = common.GetBasePath()
)

// TraceFunction traces the function and captures the metrics
func TraceFunction(f func()) {
	name := strings.ReplaceAll(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "/", "-") // Getting the name of the function

	initialGoroutines := runtime.NumGoroutine() // Capturing the initial number of goroutines
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	folderPath := fmt.Sprintf("%s/profiles", basePath)
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		log.Panicf("[MoniGo] could not create profiles directory: %v", err)
	}

	cpuProfName := fmt.Sprintf("%s_cpu.prof", name)
	cpuProfFilePath := filepath.Join(folderPath, cpuProfName)

	memProfName := fmt.Sprintf("%s_mem.prof", name)
	memProfFilePath := filepath.Join(folderPath, memProfName)

	cpuProfileFile, err := StartCPUProfile(cpuProfFilePath)
	if err != nil {
		log.Printf("[MoniGo] could not start CPU profile for function: " + name + " : Error: " + err.Error() + " will be retrying in the next iteration")
	}
	defer StopCPUProfile(cpuProfileFile)

	start := time.Now()
	f()
	elapsed := time.Since(start)

	if err := WriteHeapProfile(memProfFilePath); err != nil {
		log.Printf("[MoniGo] could not write memory profile for function: " + name + " : Error: " + err.Error() + " will be retrying in the next iteration")
	}

	runtime.ReadMemStats(&memStatsAfter)
	finalGoroutines := runtime.NumGoroutine() - initialGoroutines
	if finalGoroutines < 0 {
		finalGoroutines = 0
	}

	var memoryUsage uint64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memoryUsage = memStatsAfter.Alloc - memStatsBefore.Alloc
	}

	mu.Lock()
	defer mu.Unlock()

	functionMetrics[name] = &models.FunctionMetrics{
		FunctionLastRanAt:  start,
		CPUProfileFilePath: cpuProfFilePath,
		MemProfileFilePath: memProfFilePath,
		MemoryUsage:        memoryUsage,
		GoroutineCount:     finalGoroutines,
		ExecutionTime:      elapsed,
	}
}

// FunctionTraceDetails returns the function trace details
func FunctionTraceDetails() map[string]*models.FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()

	return functionMetrics
}

// ViewFunctionMetrics generates the function metrics
func ViewFunctionMetrics(name, reportType string, metrics *models.FunctionMetrics) models.FunctionTraceDetails {
	// Function to execute the pprof command and return the output or log an error
	executePprof := func(profileFilePath, reportType string) string {
		cmd := exec.Command("go", "tool", "pprof", "-"+reportType, profileFilePath)
		output, _ := cmd.Output()
		return string(output)
	}

	// Generating the function code stack trace for CPU profile
	codeStackView := exec.Command("go", "tool", "pprof", "-list", name, metrics.CPUProfileFilePath)
	codeStack, _ := codeStackView.Output()

	// Return the function trace details
	return models.FunctionTraceDetails{
		FunctionName: name,
		CoreProfile: models.Profiles{
			CPU: executePprof(metrics.CPUProfileFilePath, reportType),
			Mem: executePprof(metrics.MemProfileFilePath, reportType),
		},
		FunctionCodeTrace: string(codeStack),
	}
}
