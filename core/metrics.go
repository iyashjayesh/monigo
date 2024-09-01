package core

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/iyashjayesh/monigo/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	mu                      sync.Mutex
	requestCount            int64
	totalDuration           time.Duration
	functionMetrics         = make(map[string]*models.FunctionMetrics)
	serviceHealthThresholds = models.ServiceHealthThresholds{ // Default thresholds
		MaxGoroutines: models.Thresholds{
			Value:  100,
			Weight: 25,
		},
		MaxLoad: models.Thresholds{
			Value:  75.0,
			Weight: 25,
		},
		MaxMemory: models.Thresholds{
			Value:  70.0,
			Weight: 25,
		},
	}
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

func CalculateServiceHealth(metrics models.ServiceMetrics) models.ServiceHealth {
	goroutines := strconv.Itoa(int(metrics.GoRoutines))
	requests := strconv.Itoa(int(metrics.NumberOfReqServerd))
	memoryUsed := metrics.MemoryUsed
	cpuLoad := metrics.Load

	health := "Healthy"
	healthy := true
	if memoryUsed > 80.0 {
		health = "Warning: High Memory Usage"
		healthy = false
	}
	if memoryUsed > 80.0 || cpuLoad > 80.0 {
		health = "Critical: Service Under Heavy Load"
		healthy = false
	}

	// OverallHealthPercent
	overallHealthPercent := CalculateOverallHealth(&metrics)

	strToInt := func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	}

	return models.ServiceHealth{
		Goroutines:           strToInt(goroutines),
		Requests:             strToInt(requests),
		MemoryUsed:           memoryUsed,
		CPUPercent:           cpuLoad,
		OverallHealthPercent: overallHealthPercent,
		Health: models.Health{
			Healthy: healthy,
			Message: health,
		},
	}
}

// Example calculation in Go
func CalculateOverallHealth(metrics *models.ServiceMetrics) float64 {

	// Calculating the health score for each metric with a weight
	loadScore := (serviceHealthThresholds.MaxLoad.Value - metrics.Load) / serviceHealthThresholds.MaxLoad.Value * serviceHealthThresholds.MaxLoad.Weight                                                         // 25% weight
	memoryScore := (serviceHealthThresholds.MaxMemory.Value - metrics.MemoryUsed) / serviceHealthThresholds.MaxMemory.Value * serviceHealthThresholds.MaxMemory.Weight                                           // 25% weight
	goroutineScore := (float64(serviceHealthThresholds.MaxGoroutines.Value) - float64(metrics.GoRoutines)) / float64(serviceHealthThresholds.MaxGoroutines.Value) * serviceHealthThresholds.MaxGoroutines.Weight // 20% weight
	// requestScore := (float64(serviceHealthThresholds.MaxRequests.Value) - float64(requests)) / float64(serviceHealthThresholds.MaxRequests.Value) * serviceHealthThresholds.MaxRequests.Weight                   // 15% weight
	// uptimeScore := (uptime / 1440.0) * 15                                                                                                                                                                        // 15% weight, normalized for uptime (example 1440 mins = 1 day)

	// Combine into an overall health percentage
	overallHealth := loadScore + memoryScore + goroutineScore
	// + requestScore + uptimeScore

	// Ensure the health percent is within 0-100%
	if overallHealth > 100 {
		overallHealth = 100
	} else if overallHealth < 0 {
		overallHealth = 0
	}

	return overallHealth
}

func SetServiceThresholds(thresholdsValues *models.ServiceHealthThresholds) {
	serviceHealthThresholds = *thresholdsValues
}
