package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"runtime"
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
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name() // Getting the name of the function

	initialGoroutines := runtime.NumGoroutine() // Capturing the initial number of goroutines
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	log.Printf("memStatsBefore = %v\n", memStatsBefore.Alloc)

	folder := fmt.Sprintf("%s/profiles", basePath)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.Mkdir(folder, 0755); err != nil {
			fmt.Printf("Error creating profiles folder: %v\n", err)
		}
	}
	cpuProfileName := fmt.Sprintf("%s_cpu.prof", name)
	cpuProfFilePath := fmt.Sprintf("%s/%s", folder, cpuProfileName)

	cpuProfileFile, err := StartCPUProfile(cpuProfFilePath)
	if err != nil {
		fmt.Printf("Error starting CPU profile for %s: %v\n", name, err)
		return
	}
	defer StopCPUProfile(cpuProfileFile)

	memProfName := fmt.Sprintf("%s_mem.prof", name)
	memProfFilePath := fmt.Sprintf("%s/%s", folder, memProfName)

	start := time.Now()
	f()
	elapsed := time.Since(start)

	if err := WriteHeapProfile(memProfFilePath); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	// Capture final metrics
	runtime.ReadMemStats(&memStatsAfter)
	finalGoroutines := runtime.NumGoroutine() - initialGoroutines
	if finalGoroutines < 0 {
		finalGoroutines = 0
	}

	// Calculate memory usage
	var memoryUsage uint64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memoryUsage = memStatsAfter.Alloc - memStatsBefore.Alloc
	}

	log.Printf("memStatsAfter = %v\n", memStatsAfter.Alloc)
	log.Printf("memoryUsage = %v: %s\n", memoryUsage, name)

	mu.Lock()
	defer mu.Unlock()

	// Recording the metrics
	functionMetrics[name] = &models.FunctionMetrics{
		FunctionLastRanAt:  start,
		CPUProfileFilePath: cpuProfileFile.Name(),
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
