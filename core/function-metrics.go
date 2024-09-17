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

	folderPath := fmt.Sprintf("%s/profiles", basePath)

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.Mkdir(folderPath, os.ModePerm)
	}
	cpuProfileName := fmt.Sprintf("%s_cpu.prof", name)
	cpuProfFilePath := fmt.Sprintf("%s/%s", folderPath, cpuProfileName)

	cpuProfileFile, err := StartCPUProfile(cpuProfFilePath)
	if err != nil {
		log.Println("could not start CPU profile for function: ", name, " error: ", err, " It will get generated in the next run")
	}
	defer StopCPUProfile(cpuProfileFile)

	memProfName := fmt.Sprintf("%s_mem.prof", name)
	memProfFilePath := fmt.Sprintf("%s/%s", folderPath, memProfName)

	start := time.Now()
	f()
	elapsed := time.Since(start)

	if err := WriteHeapProfile(memProfFilePath); err != nil {
		log.Println("could not write memory profile: ", err, " for function: ", name, " It will get generated in the next run")
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
