package core

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	mu              sync.Mutex
	requestCount    int64
	totalDuration   time.Duration
	functionMetrics = make(map[string]*models.FunctionMetrics)
)

func RecordRequestDuration(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	requestCount++
	totalDuration += duration
}

func GetServiceMetrics() (int64, time.Duration, *runtime.MemStats) {
	mu.Lock()
	defer mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return requestCount, totalDuration, &memStats
}

func GetFunctionMetrics(functionName string) *models.FunctionMetrics {
	mu.Lock()
	defer mu.Unlock()
	return functionMetrics[functionName]
}

func GetProcessSats() models.ProcessStats {

	pid, proc, err := GetProcessDetails()
	if err != nil {
		fmt.Printf("Error fetching process information: %v\n", err)
		return models.ProcessStats{}
	}

	// Getting system and process resource usage
	sysCPUPercent, sysMemUsage, err := getSystemUsage()
	if err != nil {
		fmt.Printf("Error fetching system usage: %v\n", err)
		return models.ProcessStats{}
	}

	procCPUPercent, procMemPercent, err := getProcessUsage(proc, &sysMemUsage)
	if err != nil {
		fmt.Printf("Error fetching process usage: %v\n", err)
		return models.ProcessStats{}
	}

	totalCores, _ := cpu.Counts(false)
	totalLogicalCores := runtime.NumCPU()
	systemUsedCores := (sysCPUPercent / 100) * float64(totalLogicalCores)
	processUsedCores := (procCPUPercent / 100) * float64(totalLogicalCores)

	return models.ProcessStats{
		ProcessId:         pid,
		SysCPUPercent:     sysCPUPercent,
		ProcCPUPercent:    procCPUPercent,
		ProcMemPercent:    procMemPercent,
		TotalMemoryUsage:  sysMemUsage.TotalMemoryUsage,
		FreeMemory:        sysMemUsage.FreeMemory,
		UsedMemoryPercent: sysMemUsage.UsedPercent,
		TotalCores:        totalCores,
		TotalLogicalCores: totalLogicalCores,
		SystemUsedCores:   systemUsedCores,
		ProcessUsedCores:  processUsedCores,
	}
}

func GetProcessDetails() (int32, *process.Process, error) {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return 0, nil, err
	}
	return pid, proc, nil
}

// Fetches and returns system CPU and memory usage
func getSystemUsage() (float64, models.Memory, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, models.Memory{}, err
	}

	memUsage := models.Memory{}
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, models.Memory{}, err
	}

	memUsage = models.Memory{
		TotalMemoryUsage: float64(memInfo.Total),
		FreeMemory:       float64(memInfo.Free),
		UsedPercent:      memInfo.UsedPercent,
	}

	return cpuPercent[0], memUsage, nil
}

// Fetches and returns process CPU and memory usage
func getProcessUsage(proc *process.Process, sysMemUsage *models.Memory) (float64, float64, error) {
	procCPUPercent, err := proc.CPUPercent()
	if err != nil {
		return 0, 0, err
	}

	procMem, err := proc.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}

	if sysMemUsage.TotalMemoryUsage == 0 {
		return 0, 0, fmt.Errorf("error fetching system memory usage")
	}

	procMemPercent := float64(procMem.RSS) / sysMemUsage.TotalMemoryUsage * 100

	return procCPUPercent, procMemPercent, nil
}

func GetLocalFunctionMetrics() map[string]*models.FunctionMetrics {
	return functionMetrics
}

func GetServiceMetricsModel() models.ServiceMetrics {

	requestCount, totalDuration, memStats := GetServiceMetrics()
	serviceStat := GetProcessSats()

	// SystemUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.SystemUsedCores)
	// ProcessUsedCoresToString := fmt.Sprintf("%.2f", serviceStat.ProcessUsedCores)

	// core := ProcessUsedCoresToString + "PC / " +
	// 	SystemUsedCoresToString + "SC / " +
	// 	strconv.Itoa(serviceStat.TotalLogicalCores) + "LC / " +
	// 	strconv.Itoa(serviceStat.TotalCores) + "C"

	metrics := models.ServiceMetrics{
		Load:                   serviceStat.ProcCPUPercent,
		Cores:                  serviceStat.ProcessUsedCores,
		MemoryUsed:             float64(memStats.Alloc),
		NumberOfReqServerd:     float64(requestCount),
		GoRoutines:             float64(runtime.NumGoroutine()),
		TotalAlloc:             float64(memStats.TotalAlloc),
		MemoryAllocSys:         float64(memStats.Sys),
		HeapAlloc:              float64(memStats.HeapAlloc),
		HeapAllocSys:           float64(memStats.HeapSys),
		TotalDurationTookByAPI: totalDuration,
	}

	return metrics
}
