package core

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/iyashjayesh/monigo/models"
)

var BasePath string = GetBasePath()

func MeasureExecutionTime(name string, f func()) {

	initialGoroutines := runtime.NumGoroutine() // Capturing the initial number of goroutines
	// var memStatsAfter runtime.MemStats
	// runtime.ReadMemStats(&memStatsBefore)
	memStatsBefore := ReadMemStats()

	log.Printf("memStatsBefore = %v\n", memStatsBefore.Alloc)

	profilesFolderPath := fmt.Sprintf("%s/profiles", BasePath)

	if _, err := os.Stat(profilesFolderPath); os.IsNotExist(err) {
		os.Mkdir(profilesFolderPath, os.ModePerm)
	}

	cpuProfileName := fmt.Sprintf("%s_cpu.prof", name)
	cpuProfFilePath := fmt.Sprintf("%s/%s", profilesFolderPath, cpuProfileName)

	log.Printf("cpuProfFilePath = %s\n", cpuProfFilePath)

	cpuProfileFile, err := StartCPUProfile(cpuProfFilePath)
	if err != nil {
		log.Panicf("Error starting CPU profile for %s: %v\n", name, err)
		return
	}
	defer StopCPUProfile(cpuProfileFile)

	memProfName := fmt.Sprintf("%s_mem.prof", name)
	memProfFilePath := fmt.Sprintf("%s/%s", profilesFolderPath, memProfName)
	log.Printf("memProfFilePath = %s\n", memProfFilePath)

	start := time.Now()
	f()
	elapsed := time.Since(start)

	if err := WriteHeapProfile(memProfFilePath); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	// Capture final metrics
	// runtime.ReadMemStats(&memStatsAfter)
	memStatsAfter := ReadMemStats()
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
		FunctionLastRanAt: start,
		CPUProfile:        cpuProfileFile.Name(),
		MemoryUsage:       memoryUsage,
		GoroutineCount:    finalGoroutines,
		ExecutionTime:     elapsed,
	}
}

func GetBasePath() string {

	const monigoFolder string = "monigo"

	var path string
	appPath, _ := os.Getwd()
	if appPath == "/" {
		path = fmt.Sprintf("%s%s", appPath, monigoFolder)
	} else {
		path = fmt.Sprintf("%s/%s", appPath, monigoFolder)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	return path
}
