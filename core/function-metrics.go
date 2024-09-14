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
	functionMetrics = make(map[string]*FunctionMetrics)
	basePath        = common.GetBasePath()
)

type FunctionMetrics struct {
	FunctionLastRanAt  time.Time     `json:"function_last_ran_at"`
	CPUProfileFilePath string        `json:"cpu_profile_file_path"`
	MemProfileFilePath string        `json:"mem_profile_file_path"`
	MemoryUsage        uint64        `json:"memory_usage"`
	GoroutineCount     int           `json:"goroutine_count"`
	ExecutionTime      time.Duration `json:"execution_time"`
}

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
	functionMetrics[name] = &FunctionMetrics{
		FunctionLastRanAt:  start,
		CPUProfileFilePath: cpuProfileFile.Name(),
		MemProfileFilePath: memProfFilePath,
		MemoryUsage:        memoryUsage,
		GoroutineCount:     finalGoroutines,
		ExecutionTime:      elapsed,
	}
}

func FunctionTraceDetails() map[string]*FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()

	return functionMetrics
}

func ViewFunctionMetrics(name string, reportType string, metrics *FunctionMetrics) models.FunctionTraceDetails {

	// https://github.com/google/pprof/blob/main/doc/README.md#text-reports
	// 	Text reports
	// pprof text reports show the location hierarchy in text format.

	// -text: Prints the location entries, one per line, including the flat and cum values.
	// -tree: Prints each location entry with its predecessors and successors.
	// -peek= regex: Print the location entry with all its predecessors and successors, without trimming any entries.
	// -traces: Prints each sample with a location per line.

	log.Println("Metrics: ", metrics)
	cmdCpu := exec.Command("go", "tool", "pprof", "-"+reportType, metrics.CPUProfileFilePath)
	cpu, err := cmdCpu.Output()
	if err != nil {
		log.Println("failed to generate cpu profile: %v\n", err)
	}

	cmdMem := exec.Command("go", "tool", "pprof", "-"+reportType, metrics.MemProfileFilePath)
	mem, err := cmdMem.Output()
	if err != nil {
		log.Println("failed to generate mem profile: %v\n", err)
	}

	// go tool pprof -list main.highMemoryUsage  monigo/profiles/main.highMemoryUsage_cpu.prof
	codeStackView := exec.Command("go", "tool", "pprof", "-list", name, metrics.CPUProfileFilePath)
	codeStack, err := codeStackView.Output()
	if err != nil {
		log.Println("failed to generate cpu code stack view: %v\n", err)
	}

	// codememStackView := exec.Command("go", "tool", "pprof", "-list", name, metrics.MemProfileFilePath)
	// log.Println("codememStackView: ", codememStackView)
	// codeMemStack, err := codememStackView.Output()
	// if err != nil {
	// 	log.Println("failed to generate mem code stack view: %v\n", err)
	// }

	return models.FunctionTraceDetails{
		FunctionName: name,
		CoreProfile: models.Profiles{
			CPU: string(cpu),
			Mem: string(mem),
		},
		FunctionCodeTrace: string(codeStack),
	}
}
